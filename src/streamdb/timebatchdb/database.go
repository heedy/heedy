package timebatchdb

//Package timebatchdb is a time series Database built to handle extremely fast messaging as well as
//pretty large amounts of data.

import (
	"database/sql"
	"errors"

	log "github.com/Sirupsen/logrus"
)

var (

	//ErrorIndexMismatch is thrown when the index in RedisCache does not match index in SqlStore. It means that there seems to be data missing from the database
	ErrorIndexMismatch = errors.New("Database internal index mismatch")
	//ErrorUserFail is returned when the data ranges requested are either both the same, or are somehow invalid.
	ErrorUserFail = errors.New("U FAIL @ LYFE (Check your data range)")
)

//The databaseRange structure conforms to the DataRange interface. It represents a range of data from a given stream. The structure is used internally.
type databaseRange struct {
	db    *Database
	dr    DataRange
	index uint64
	key   string
}

//Init doesn't actually do anything - it just conforms to the DataRange interface
func (d *databaseRange) Init() error {
	return nil
}

//Close shuts stuff down and releases resources
func (d *databaseRange) Close() {
	d.dr.Close()
}

//Next is an iterator which returns the next datapoint in the stream on each call.
func (d *databaseRange) Next() (*Datapoint, error) {
	dp, err := d.dr.Next()
	//If there is an explicit error - or if there was a datapoint returned, just go with it
	if err != nil || dp != nil {
		d.index++
		return dp, err
	}

	//If the program got here, it means that the DataRange is empty. We therefore check where to look for the next
	//data index.
	startindex, cachelength, err := d.db.cache.GetIndices(d.key)
	if err != nil {
		return nil, err
	}

	if startindex > d.index {
		//We look to the sql database for the next datapoint
		d.dr, startindex, err = d.db.store.GetByIndex(d.key, d.index)
		if err != nil {
			return nil, err
		}
		if startindex != d.index {
			return nil, ErrorIndexMismatch
		}

		//The sqlrange can be empty in certain cases. We therefore check for database corruption
		dp, err = d.dr.Next()
		if err != nil || dp != nil {
			d.index++
			return dp, err
		}
		//If it gets here there was no error and dp was nil. This means that the database is corrupted
		return nil, ErrorDatabaseCorrupted

	} else if d.index >= startindex+cachelength {
		return nil, nil //The index is out of bounds - return nil
	} else {
		//The index should be in the cache
		d.dr, startindex, err = d.db.cache.GetByIndex(d.key, d.index)
		if err != nil {
			return nil, err
		}
		if startindex > d.index { //This means that the data we wanted was just written to the database. So redo the procedure.
			d.dr = EmptyRange{}
		} else if startindex != d.index {
			return nil, ErrorIndexMismatch
		}
	}
	return d.Next()
}

//The Database object handles all querying and inserting into TimebatchDB
type Database struct {
	cache     *RedisCache
	store     *SqlStore
	batchsize int
}

//Close releases all resources that TimebatchDB has taken
func (d *Database) Close() {
	d.cache.Close()
	d.store.Close()
}

//Delete the given key from the database
func (d *Database) Delete(key string) error {
	err := d.cache.Delete(key)
	if err != nil {
		return err
	}
	return d.store.Delete(key)
}

//DeletePrefix deletes all keys which start with prefix from the database
func (d *Database) DeletePrefix(prefix string) error {
	_, err := d.cache.DeletePrefix(prefix)
	if err != nil {
		return err
	}
	return d.store.DeletePrefix(prefix)
}

//Len gets the total number of datapoints for the given key
func (d *Database) Len(key string) (uint64, error) {
	return d.cache.EndIndex(key)
}

//GetIndexRange gets the given range of index values for the given key from the database
func (d *Database) GetIndexRange(key string, i1 uint64, i2 uint64) (DataRange, error) {
	if i1 >= i2 {
		return EmptyRange{}, ErrorUserFail
	}
	return NewNumRange(&databaseRange{d, EmptyRange{}, i1, key}, i2-i1), nil
}

//GetTimeRange gets the given time range of values given a key from the database
func (d *Database) GetTimeRange(key string, t1 int64, t2 int64) (DataRange, error) {
	if t1 >= t2 {
		return EmptyRange{}, ErrorUserFail
	}
	//We have to be more clever here - we will need to initialize the databaseRange in
	//the middle of a query so that it can use indices instead of timestamps.
	startTime, err := d.cache.GetStartTime(key)
	if err != nil {
		return EmptyRange{}, err
	}
	if startTime <= t1 {
		//The whole data range is within the cache... Make sure that it isn't outside of all data (no need to waste CPU cycles)
		et, err := d.cache.GetEndTime(key)
		if err != nil || et <= t1 {
			return EmptyRange{}, err
		}

		//Alright, attempt to get the data from the cache
		datapointarray, startIndex, err := d.cache.Get(key)
		if err != nil {
			return EmptyRange{}, err
		}

		//Now make sure that the timestamp matches what we got from the cache
		if datapointarray.Len() > 0 && datapointarray.Datapoints[0].Timestamp() == startTime {
			return NewTimeRange(&databaseRange{d, datapointarray, startIndex, key}, t1, t2), nil
		}
		//If we are here, then some shit happened. We're not going to deal with this BS, so we just query the sqlstore.
	}

	dataRange, startIndex, err := d.store.GetByTime(key, t1)
	if err != nil {
		return EmptyRange{}, err
	}
	return NewTimeRange(&databaseRange{d, dataRange, startIndex, key}, t1, t2), nil
}

//Insert the given datapoint array to the stream given at key.
func (d *Database) Insert(key string, datapointarray *DatapointArray) error {
	return d.cache.Insert(key, datapointarray, d.batchsize)
}

//WriteDatabaseIteration runs one iteration of WriteDatabase - it blocks until a batch is ready, processes the batch, and returns.
func (d *Database) WriteDatabaseIteration() (err error) {
	key, err := d.cache.BatchWait()
	if err != nil {
		return err
	}
	//Now we compare the end index in redis to that of the sql database
	storeEndIndex, err := d.store.GetEndIndex(key)
	if err != nil {
		d.cache.BatchPush(key) //Try to make future recovery possible - repush the current key
		return err
	}

	datapointarray, cacheStartIndex, err := d.cache.BatchGet(key, d.batchsize)
	if err == ErrorRedisWrongSize || datapointarray.Len() < d.batchsize { //If WrongSize, it means that the key was pushed needlessly - ignore the key
		log.Warningf("TimebatchDB:WriteDatabase:IGNORING: Got small batch: key=%v #=%v", key, datapointarray.Len())
	} else if err != nil {
		d.cache.BatchPush(key)
		return err
	} else {

		if storeEndIndex == cacheStartIndex {
			//Looks like all is well - write the datapoints to database
			err = d.store.Insert(key, cacheStartIndex, datapointarray)
			if err != nil {
				d.cache.BatchPush(key)
				return err
			}

			err = d.cache.BatchRemove(key, datapointarray.Len())
			if err != nil {
				d.cache.BatchPush(key)
				return err
			}
			log.Debugf("TimebatchDB:WriteDatabase: Wrote Key=%s I=%v #=%v", key, cacheStartIndex, datapointarray.Len())
		} else if storeEndIndex < cacheStartIndex {
			d.cache.BatchPush(key)
			return ErrorIndexMismatch //O shit. This breaks the database.
		} else {
			//This is the unusual situation where there cacheStartIndex < storeEndIndex. This can happen if BatchRemove fails
			//in an earlier iteration, or if there is a concurrent database connection dealing with the same key
			//which is in the middle of inserting data. We don't know which it is... so we assume that BatchRemove failed,
			//and the functino is not running concurrently
			//TODO: There should probably be some code here that takes care of this situation is a non-bad way for concurrency
			log.Warningln("TimebatchDB:WriteDatabase: cache_start < store_end :", key)
			/*
			   err = d.cache.BatchRemove(key,d.batchsize)
			   if (err!=nil) {
			       break
			   }
			*/
		}
	}
	return nil
}

//WriteDatabase blocks indefinitely, reading from the data cache and writing to the SQL database.
//If running as part of a larger application, you probably want this as a goroutine or as its own process.
//In order for the database to function properly, there needs to be an instance of this function running somewhere in the background
//Please note that while I think that the function will function concurrently, this was not stress-tested for concurrency issues.
func (d *Database) WriteDatabase() (err error) {
	log.Debugln("TimebatchDB:WriteDatabase: running")
	for err == nil {
		err = d.WriteDatabaseIteration()
	}
	log.Errorln("TimebatchDB:WriteDatabase:", err)
	return err
}

//WriteLoop is just like WriteDatabase - but it just keeps restarting the writer on error
func (d *Database) WriteLoop() {
	for {
		d.WriteDatabase()
	}
}

//Open the database given the necessary inputs. Last error input is for error-chaining. use nil if not interested.
func Open(sdb *sql.DB, sqldriver string, redisurl string, batchsize int, err error) (*Database, error) {
	store, err := OpenSqlStore(sdb, sqldriver, err)
	cache, err := OpenRedisCache(redisurl, err)
	if err != nil {
		return nil, err
	}
	return &Database{cache, store, batchsize}, nil
}
