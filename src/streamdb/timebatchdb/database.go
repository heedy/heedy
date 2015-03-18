//TimeBatchDB is a time series Database built to handle extremely fast messaging as well as
//pretty large amounts of data.
package timebatchdb

import (
	"database/sql"
	"errors"
	"log"
)

var (
	ERROR_UNORDERED      = errors.New("Datapoints not ordered by timestamp")
	ERROR_TIMESTAMP      = errors.New("A datapoint with a greater or equal timestamp already exists for the stream")
	ERROR_INDEX_MISMATCH = errors.New("Database internal index mismatch - possible data loss!")
	ERROR_USERFAIL       = errors.New("U FAIL @ LYFE (Check your data range)")
)

//This structure conforms to the DataRange interface
type DatabaseRange struct {
	db    *Database
	dr    DataRange
	index uint64
	key   string
}

//This doesn't actually do anything - it just conforms to the DataRange interface
func (d *DatabaseRange) Init() error {
	return nil
}

//Shuts stuff down and releases resources
func (d *DatabaseRange) Close() {
	d.dr.Close()
}

func (d *DatabaseRange) Next() (*Datapoint, error) {
	dp, err := d.dr.Next()
	//If there is an explicit error - or if there was a datapoint returned, just go with it
	if err != nil || dp != nil {
		d.index++
		return dp, err
	}

	//If the program got here, it means that the DataRange is empty. We therefore check where to look for the next
	//data index.
	si, cl, err := d.db.rc.GetIndices(d.key)
	if err != nil {
		return nil, err
	}

	if si > d.index {
		//We look to the sql database for the next datapoint
		d.dr, si, err = d.db.ss.GetByIndex(d.key, d.index)
		if err != nil {
			return nil, err
		}
		if si != d.index {
			return nil, ERROR_INDEX_MISMATCH
		}
	} else if d.index >= si+cl {
		return nil, nil //The index is out of bounds - return nil
	} else {
		//The index should be in the cache
		d.dr, si, err = d.db.rc.GetByIndex(d.key, d.index)
		if err != nil {
			return nil, err
		}
		if si > d.index { //This means that the data we wanted was just written to the database. So redo the procedure.
			d.dr = EmptyRange{}
		} else if si != d.index {
			return nil, ERROR_INDEX_MISMATCH
		}
	}
	return d.Next()
}

//The Database object handles all querying and inserting into TimebatchDB
type Database struct {
	rc        *RedisCache
	ss        *SqlStore
	batchsize int
}

//Release all resources that TimebatchDB has taken
func (d *Database) Close() {
	d.rc.Close()
	d.ss.Close()
}

//Deletes the given key from the database
func (d *Database) Delete(key string) error {
	err := d.rc.Delete(key)
	if err != nil {
		return err
	}
	return d.ss.Delete(key)
}

//Gets the total number of datapoints for the given key
func (d *Database) Len(key string) (uint64, error) {
	return d.rc.EndIndex(key)
}

//Gets the given range of index values for the given key from the database
func (d *Database) GetIndexRange(key string, i1 uint64, i2 uint64) (DataRange, error) {
	if i1 >= i2 {
		return EmptyRange{}, ERROR_USERFAIL
	}
	return NewNumRange(&DatabaseRange{d, EmptyRange{}, i1, key}, i2-i1), nil
}

//Gets the given time range of values given a key from the database
func (d *Database) GetTimeRange(key string, t1 int64, t2 int64) (DataRange, error) {
	if t1 >= t2 {
		return EmptyRange{}, ERROR_USERFAIL
	}
	//We have to be more clever here - we will need to initialize the DatabaseRange in
	//the middle of a query so that it can use indices instead of timestamps.
	st, err := d.rc.GetStartTime(key)
	if err != nil {
		return EmptyRange{}, err
	}
	if st <= t1 {
		//The whole data range is within the cache... Make sure that it isn't outside of all data (no need to waste CPU cycles)
		et, err := d.rc.GetEndTime(key)
		if err != nil || et <= t1 {
			return EmptyRange{}, err
		}

		//Alright, attempt to get the data from the cache
		dpa, si, err := d.rc.Get(key)
		if err != nil {
			return EmptyRange{}, err
		}

		//Now make sure that the timestamp matches what we got from the cache
		if dpa.Len() > 0 && dpa.Datapoints[0].Timestamp() == st {
			return NewTimeRange(&DatabaseRange{d, dpa, si, key}, t1, t2), nil
		}
		//If we are here, then some shit happened. We're not going to deal with this BS, so we just query the sqlstore.
	}

	sr, si, err := d.ss.GetByTime(key, t1)
	if err != nil {
		return EmptyRange{}, err
	}
	return NewTimeRange(&DatabaseRange{d, sr, si, key}, t1, t2), nil
}

//Inserts the given datapoint array to the stream given at key.
func (d *Database) Insert(key string, dpa *DatapointArray) error {
	if !dpa.IsTimestampOrdered() || dpa.Len() == 0 {
		return ERROR_UNORDERED
	}
	et, clen, err := d.rc.GetEndTimeAndCacheLength(key)
	if et >= dpa.Datapoints[0].Timestamp() {
		return ERROR_TIMESTAMP
	} else if err != nil {
		return err
	}
	//If the batch size was exceeded on this insert, add it to the queue (this only adds
	//to the write queue on change in division, so that a batch is only written once to the database)
	batchnum := (clen+dpa.Len())/d.batchsize - clen/d.batchsize
	return d.rc.InsertAndBatchPush(key, dpa, batchnum)
}

//This runs one iteration of WriteDatabase - it blocks until a batch is ready, processes the batch, and returns.
func (d *Database) WriteDatabaseIteration() (err error) {
	key, err := d.rc.BatchWait()
	if err != nil {
		return err
	}
	//Now we compare the end index in redis to that of the sql database
	s_end, err := d.ss.GetEndIndex(key)
	if err != nil {
		d.rc.BatchPush(key) //Try to make future recovery possible - repush the current key
		return err
	}

	dpa, r_start, err := d.rc.BatchGet(key, d.batchsize)
	if err == ERROR_REDIS_WRONGSIZE { //If WrongSize, it means that the key was pushed needlessly - ignore the key
		log.Println("TimebatchDB:WriteDatabase:WARNING: Got batch where there is none:", key)
	} else if err != nil {
		d.rc.BatchPush(key)
		return err
	} else {

		if s_end == r_start {
			//Looks like all is well - write the datapoint to database
			err = d.ss.Insert(key, r_start, dpa)
			if err != nil {
				d.rc.BatchPush(key)
				return err
			}

			err = d.rc.BatchRemove(key, d.batchsize)
			if err != nil {
				d.rc.BatchPush(key)
				return err
			}
			log.Printf("TimebatchDB:WriteDatabase: Wrote Key=%s I=%v #=%v\n", key, r_start, dpa.Len())
		} else if s_end < r_start {
			d.rc.BatchPush(key)
			return ERROR_INDEX_MISMATCH //O shit. This breaks the database.
		} else {
			//This is the unusual situation where there r_start < s_end. This can happen if BatchRemove fails
			//in an earlier iteration, or if there is a concurrent database connection dealing with the same key
			//which is in the middle of inserting data. We don't know which it is... so we assume that BatchRemove failed,
			//and the functino is not running concurrently
			//TODO: There should probably be some code here that takes care of this situation is a non-bad way for concurrency
			log.Println("TimebatchDB:WriteDatabase:WARNING: cache_start < store_end :", key)
			/*
			   err = d.rc.BatchRemove(key,d.batchsize)
			   if (err!=nil) {
			       break
			   }
			*/
		}
	}
	return nil
}

//This function blocks indefinitely, reading from the data cache and writing to the SQL database.
//If running as part of a larger application, you probably want this as a goroutine or as its own process.
//In order for the database to function properly, there needs to be an instance of this function running somewhere in the background
//Please note that while I think that the function will function concurrently, this was not stress-tested for concurrency issues.
func (d *Database) WriteDatabase() (err error) {
	log.Println("TimebatchDB:WriteDatabase:RUNNING")
	for err == nil {
		err = d.WriteDatabaseIteration()
	}
	log.Println("TimebatchDB:WriteDatabase:ERROR ", err)
	return err
}

//This is just like WriteDatabase - but it just keeps restarting the writer on error
func (d *Database) WriteLoop() {
	for {
		d.WriteDatabase()
	}
}

//Opens the database given the necessary inputs. Last error input is for error-chaining. use nil if not interested.
func Open(sdb *sql.DB, sqldriver string, redisurl string, batchsize int, err error) (*Database, error) {
	ss, err := OpenSqlStore(sdb, sqldriver, err)
	rc, err := OpenRedisCache(redisurl, err)
	if err != nil {
		return nil, err
	}
	return &Database{rc, ss, batchsize}, nil
}
