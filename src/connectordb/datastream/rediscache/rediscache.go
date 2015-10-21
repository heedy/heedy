package rediscache

import (
	"connectordb/datastream"
	"strconv"
)

//RedisCache reads batches from a single-instance redis server
type RedisCache struct {
	*RedisConnection
}

//StreamLength returns the length of a stream
func (r RedisCache) StreamLength(deviceID, streamID int64, substream string) (int64, error) {
	return r.RedisConnection.StreamLength(
		strconv.FormatInt(deviceID, 36),
		strconv.FormatInt(streamID, 36),
		substream)
}

//Insert datapoints into the redis cache
func (r RedisCache) Insert(deviceID, streamID int64, substream string, dpa datastream.DatapointArray, restamp bool) (int64, error) {
	return r.RedisConnection.Insert("BATCHLIST",
		strconv.FormatInt(deviceID, 36),
		strconv.FormatInt(streamID, 36),
		substream,
		dpa,
		restamp)
}

//DeleteDevice removes a device from the redis cache
func (r RedisCache) DeleteDevice(deviceID int64) error {
	return r.DeleteHash(strconv.FormatInt(deviceID, 36))
}

//DeleteStream removes a stream from the redis cache
func (r RedisCache) DeleteStream(deviceID, streamID int64) error {
	return r.RedisConnection.DeleteStream(strconv.FormatInt(deviceID, 36), strconv.FormatInt(streamID, 36))
}

//DeleteSubstream removes a substream fro mthe redis cache
func (r RedisCache) DeleteSubstream(deviceID, streamID int64, substream string) error {
	return r.RedisConnection.DeleteSubstream(strconv.FormatInt(deviceID, 36), strconv.FormatInt(streamID, 36), substream)
}

//ReadProcessingQueue reads all the batches in the processing queue
func (r RedisCache) ReadProcessingQueue() ([]datastream.Batch, error) {
	bstrings, err := r.GetList("BATCHPROCESSING")
	if err != nil || len(bstrings) == 0 {
		return nil, err
	}
	barray := make([]datastream.Batch, len(bstrings))
	for i := 0; i < len(bstrings); i++ {
		v, err := r.ReadBatch(bstrings[i])
		if err != nil {
			return nil, err
		}
		//The ordering is reversed because batches from the same stream need to come in in order of index
		//for trim to work properly
		barray[len(bstrings)-1-i] = *v
	}
	return barray, nil
}

//ReadBatches reads the given number of batches from the batch list
func (r RedisCache) ReadBatches(batchnumber int) ([]datastream.Batch, error) {
	barray := make([]datastream.Batch, batchnumber)

	for i := 0; i < batchnumber; i++ {
		batchstring, err := r.NextBatch("BATCHLIST", "BATCHPROCESSING")
		if err != nil {
			return nil, err
		}
		v, err := r.ReadBatch(batchstring)
		if err != nil {
			return nil, err
		}
		barray[i] = *v
	}
	return barray, nil
}

//ReadRange reads the given range from the given stream
func (r RedisCache) ReadRange(deviceID, streamID int64, substream string, i1, i2 int64) (datastream.DatapointArray, int64, int64, error) {
	return r.Range(
		strconv.FormatInt(deviceID, 36),
		strconv.FormatInt(streamID, 36),
		substream, i1, i2)
}

//ClearBatches clears the batches that are listed as "processing", and removes the associated
//datapoints from their streams
func (r RedisCache) ClearBatches(b []datastream.Batch) error {

	//In the future more checks should be made to ensure that we are not losing anything
	bstrings, err := r.GetList("BATCHPROCESSING")
	if err != nil {
		return err
	}
	if len(b) != len(bstrings) {
		return ErrWTF
	}

	//Now trim the streams in the batches
	for i := 0; i < len(b); i++ {
		err = r.TrimStream(b[i].Device, b[i].Stream, b[i].Substream, b[i].EndIndex())
		if err != nil {
			return err
		}
	}
	return r.DeleteKey("BATCHPROCESSING")
}
