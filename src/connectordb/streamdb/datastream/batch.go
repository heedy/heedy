package datastream

import "strconv"

//Batch is a struct that encodes a batch of data coming from the cache
type Batch struct {
	Stream    string
	Substream string
	Device    string

	StartIndex int64

	//The data associated with the batch
	Data DatapointArray
}

//GetStreamID returns the stream as an int encoded by SetStreamID
func (b *Batch) GetStreamID() (int64, error) {
	return strconv.ParseInt(b.Stream, 36, 64)
}

//SetStreamID sets the stream string given the stream ID
func (b *Batch) SetStreamID(streamID int64) {
	b.Stream = strconv.FormatInt(streamID, 36)
}

//GetDeviceID returns the stream as an int encoded by SetStreamID
func (b *Batch) GetDeviceID() (int64, error) {
	return strconv.ParseInt(b.Device, 36, 64)
}

//SetDeviceID sets the stream string given the stream ID
func (b *Batch) SetDeviceID(deviceID int64) {
	b.Device = strconv.FormatInt(deviceID, 36)
}

//EndIndex returns the end index of the batch
func (b *Batch) EndIndex() int64 {
	return b.StartIndex + int64(len(b.Data))
}
