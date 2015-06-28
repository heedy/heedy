package datastream

type DataStream struct {
}

//Insert inserts the given datapoint array into the stream
func (ds *DataStream) Insert(stream int64, substream string, dpa []Datapoint) error {
	return nil
}

//InsertRestamp inserts the given datapoint array into the stream - and it restamps the timestamp
//on the datapoint to the insert time (this means there can be no timestamp errors)
func (ds *DataStream) InsertRestamp(stream int64, substream string, dpa []Datapoint) error {
	return nil
}
