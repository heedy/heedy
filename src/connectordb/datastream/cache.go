/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package datastream

//Cache is an interface that caches datapoints for all the streams until there are enoguh in memory to form a batch
//of data
type Cache interface {
	StreamLength(deviceID int64, streamID int64, substream string) (int64, error)
	DeviceSize(deviceID int64) (int64, error)
	StreamSize(deviceID, streamID int64, substream string) (int64, error)
	Insert(deviceID, streamID int64, substream string, dpa DatapointArray, restamp bool, maxDeviceSize int64, maxStreamSize int64) (int64, error)
	DeleteDevice(deviceID int64) error
	DeleteStream(deviceID, streamID int64) error
	DeleteSubstream(deviceID, streamID int64, substream string) error
	ReadProcessingQueue() ([]Batch, error)
	ReadBatches(batchnumber int) ([]Batch, error)
	ReadRange(deviceID, streamID int64, substream string, i1, i2 int64) (DatapointArray, int64, int64, error)
	ClearBatches(b []Batch) error
	Close() error
	Clear() error
}
