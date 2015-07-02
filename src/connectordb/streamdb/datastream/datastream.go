package datastream

import "errors"

var (
	//ErrTimestampOrder is thrown when out of order tiemstamps are detected
	ErrTimestampOrder = errors.New("The datapoints must be ordered by increasing timestamp")
)

/*
//stringStream returns a string representatino of the streamID
func stringStream(stream int64) string {
	return strconv.FormatInt(stream, 32)
}

//DataStream is how the database extracts data from a stream. It is the main object in datastream
type DataStream struct {
	redis *RedisConnection
	sqls  *SqlStore

	batchsize int64
}

//OpenDataStream does just that - it opens the DataStream
func OpenDataStream(sd *sql.DB, o *Options) (ds *DataStream, err error) {
	redis, err := NewRedisConnection(o)
	if err != nil {
		return nil, err
	}
	sqls, err := OpenSqlStore(sd)
	if err != nil {
		return nil, err
	}
	return &DataStream{redis, sqls, int64(o.BatchSize)}, nil
}

//Close releases all resources held by the DataStream. It does NOT close open DataRanges
func (ds *DataStream) Close() {
	ds.redis.Close()
	ds.sqls.Close()
}

//Clear removes all data held in the database. Only to be used for testing purposes!
func (ds *DataStream) Clear() {
	ds.redis.Clear()
	ds.sqls.Clear()
}

//DeleteStream deletes an entire stream from the database
func (ds *DataStream) DeleteStream(stream int64) error {
	err := ds.redis.DeleteStream(stringStream(stream))
	if err != nil {
		return err
	}
	return ds.sqls.DeleteStream(stream)
}

//DeleteSubstream deletes the substream from the database
func (ds *DataStream) DeleteSubstream(stream int64, substream string) error {
	err := ds.redis.DeleteSubstream(stringStream(stream), substream)
	if err != nil {
		return err
	}
	return ds.sqls.DeleteSubstream(stream, substream)
}

//StreamLength returns the length of the stream
func (ds *DataStream) StreamLength(stream int64, substream string) (int64, error) {
	return ds.redis.StreamLength(stringStream(stream), substream)
}

//Insert inserts the given datapoint array into the stream, with the option to restamp the data
//on insert if it has timestamps below the range of already-inserted data. Restamoing allows an insert to always succeed
func (ds *DataStream) Insert(stream int64, substream string, dpa DatapointArray, restamp bool) error {
	//NOTE: Assuming that the datapointarray is correctly ordered by timestamp.

	//Insert the data into redis
	slength, err := ds.redis.Insert(stringStream(stream), substream, dpa, restamp)
	if err != nil {
		return err
	}

	//Now that the datapoints are inserted, check to see if we can write a new batch
	//to the long term storage
	batchnumber := slength/ds.batchsize - (slength-int64(dpa.Length()))/ds.batchsize
	if batchnumber > 0 {
		//There are batches to write!

	}

	return nil
}

/*
//IRange returns a DataRange of datapoints which are in the given range of indices.
//Indices can be python-like, meaning i1 and i2 negative mean "from the end", and i2=0
//means to the end.
func (ds *DataStream) IRange(stream int64, substream string, i1 int64, i2 int64) (dr DataRange, err error) {
	dpa, i1, i2, err := ds.redis.Range(stringStream(stream), substream, i1, i2)
	if err != nil {
		return EmptyRange{}, err
	}
	if dpa != nil {
		//Aww yes, the entire range was in redis
		return NewDatapointArrayRange(dpa), nil
	}

	//We query the datastore by index
	return nil,
}
*/
