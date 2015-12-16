/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package datastream

import (
	"database/sql"
	"errors"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrTimestampOrder is thrown when out of order tiemstamps are detected
	ErrTimestampOrder = errors.New("The datapoints must be ordered by increasing timestamp")
)

//DataStream is how the database extracts data from a stream. It is the main object in datastream
type DataStream struct {
	cache Cache
	sqls  *SqlStore

	//ChunkSize is the number of batches to write to postgres in one transaction.
	ChunkSize int
}

//OpenDataStream does just that - it opens the DataStream
func OpenDataStream(c Cache, sd *sql.DB, chunksize int) (ds *DataStream, err error) {
	sqls, err := OpenSqlStore(sd)
	if err != nil {
		return nil, err
	}
	return &DataStream{c, sqls, chunksize}, nil
}

//Close releases all resources held by the DataStream. It does NOT close open ExtendedDataRanges
func (ds *DataStream) Close() {
	ds.cache.Close()
	ds.sqls.Close()
}

//Clear removes all data held in the database. Only to be used for testing purposes!
func (ds *DataStream) Clear() {
	ds.cache.Clear()
	ds.sqls.Clear()
}

//DeleteDevice deletes a device from the cache (propagation takes care of deleting it from the sqlstore)
func (ds *DataStream) DeleteDevice(deviceID int64) error {
	return ds.cache.DeleteDevice(deviceID)
}

//DeleteStream deletes an entire stream from the database
func (ds *DataStream) DeleteStream(deviceID, streamID int64) error {
	err := ds.cache.DeleteStream(deviceID, streamID)
	if err != nil {
		return err
	}
	return ds.sqls.DeleteStream(streamID)
}

//DeleteSubstream deletes the substream from the database
func (ds *DataStream) DeleteSubstream(deviceID, streamID int64, substream string) error {
	err := ds.cache.DeleteSubstream(deviceID, streamID, substream)
	if err != nil {
		return err
	}
	return ds.sqls.DeleteSubstream(streamID, substream)
}

//StreamLength returns the length of the stream
func (ds *DataStream) StreamLength(deviceID, streamID int64, substream string) (int64, error) {
	return ds.cache.StreamLength(deviceID, streamID, substream)
}

//Insert inserts the given datapoint array into the stream, with the option to restamp the data
//on insert if it has timestamps below the range of already-inserted data. Restamping allows an insert to always succeed
func (ds *DataStream) Insert(deviceID, streamID int64, substream string, dpa DatapointArray, restamp bool) (int64, error) {
	if !dpa.IsTimestampOrdered() {
		return 0, ErrTimestampOrder
	}
	return ds.cache.Insert(deviceID, streamID, substream, dpa, restamp)
}

//WriteChunk takes a chunk of batches and writes it to the sql store
func (ds *DataStream) WriteChunk() error {
	b, err := ds.cache.ReadBatches(ds.ChunkSize)
	if err != nil {
		return err
	}
	if err = ds.sqls.WriteBatches(b); err != nil {
		return err
	}
	return ds.cache.ClearBatches(b)
}

//WriteQueue writes the queue of leftover data that might have been half-processed
func (ds *DataStream) WriteQueue() error {
	log.Debug("DBWriter: Checking write queue...")
	b, err := ds.cache.ReadProcessingQueue()
	if err != nil {
		return err
	}
	if len(b) > 0 {
		if err = ds.sqls.WriteBatches(b); err != nil {
			return err
		}
	}
	return ds.cache.ClearBatches(b)
}

//RunWriter runs writer in a loop FOREVAAAARRRR
func (ds *DataStream) RunWriter() error {
	log.Debug("Starting Database Writer")
	err := ds.WriteQueue()
	log.Debug("Running DBWriter")
	for err == nil {
		err = ds.WriteChunk()
	}
	//This error display interferes with benchmarks which is annoying.
	log.Errorf("DBWriter error: %v", err.Error())
	return err
}

//IRange returns a ExtendedDataRange of datapoints which are in the given range of indices.
//Indices can be python-like, meaning i1 and i2 negative mean "from the end", and i2=0
//means to the end.
func (ds *DataStream) IRange(device int64, stream int64, substream string, i1 int64, i2 int64) (dr ExtendedDataRange, err error) {
	dpa, i1, i2, err := ds.cache.ReadRange(device, stream, substream, i1, i2)
	if err != nil || i1 == i2 {
		return EmptyRange{}, err
	}
	if dpa != nil {
		//Aww yes, the entire range was in redis
		return NewDatapointArrayRange(dpa, i1), nil
	}

	//At least part of the range was in sql. So query sql with it, and return the StreamRange
	//object with the correct initialization
	sqlr, i1, err := ds.sqls.GetByIndex(stream, substream, i1)

	return NewNumRange(&StreamRange{
		ds:        ds,
		dr:        sqlr,
		index:     i1,
		deviceID:  device,
		streamID:  stream,
		substream: substream,
	}, i2-i1), err
}

//TRange returns a ExtendedDataRange of datapoints which are in the given range of timestamp.
func (ds *DataStream) TRange(device int64, stream int64, substream string, t1, t2 float64) (dr ExtendedDataRange, err error) {
	//TRange works a bit differently from IRange, since time ranges go straight to postgres
	sqlr, startindex, err := ds.sqls.GetByTime(stream, substream, t1)

	if err != nil {
		return EmptyRange{}, err
	}

	return NewTimeRange(&StreamRange{
		ds:        ds,
		dr:        sqlr,
		index:     startindex,
		deviceID:  device,
		streamID:  stream,
		substream: substream,
	}, t1, t2)
}

//GetTimeIndex returns the corresponding index of data given a timestamp
func (ds *DataStream) GetTimeIndex(device int64, stream int64, substream string, t float64) (int64, error) {
	dr, err := ds.TRange(device, stream, substream, t, 0.0)
	if err != nil {
		return 0, err
	}
	return dr.Index(), nil
}

//TimePlusIndexRange returns a range starting at the given time, offset by the given index (+ or -),
//	The range is to the end of the entire stream (ie, just close it when you don't need further data)
//TODO: This function can be made much more efficient with a bit of cleverness regarding the underlying
//	ExtendedDataRanges
func (ds *DataStream) TimePlusIndexRange(device int64, stream int64, substream string, t1, t2 float64, i int64) (ExtendedDataRange, error) {
	//First off, we get the TRange
	dr, err := ds.TRange(device, stream, substream, t1, t2)
	if err != nil {
		return nil, err
	}
	if i < 0 || i > int64(ds.ChunkSize*2) {
		//In this case, we will need to query the database again - this time by index
		curindex := dr.Index()
		dr.Close()

		i = curindex + i
		if i < 0 {
			//While this could return an error, most use cases just want to start from the beginning of data
			//	in particular the dataset interpolators
			i = 0
		}
		irng, err := ds.IRange(device, stream, substream, i, 0)
		if err != nil {
			return irng, err
		}
		return NewTimeRange(irng, -999999, t2)
	}

	//No need to query again, extract the new datarange from current one

	//Currently, the only possibility is if i is >=0 and small, so we just iterate through
	for j := int64(0); j < i; j++ {
		_, err := dr.Next()
		if err != nil {
			dr.Close()
			return nil, err
		}
	}

	return dr, nil
}
