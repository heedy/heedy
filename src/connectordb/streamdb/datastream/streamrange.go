package datastream

//StreamRange is a DataRange that combines the redis and sql data into one coherent stream
type StreamRange struct {
	ds *DataStream
	dr DataRange

	index      int64
	streamID   int64
	streamName string
	substream  string
}

//Close the StreamRange
func (d *StreamRange) Close() {
	if d.dr != nil {
		d.dr.Close()
	}
}

//NextArray returns the next datapoint array from the stream
func (d *StreamRange) NextArray() (*DatapointArray, error) {
	return nil, nil
}

//Next returns the next datapoint
func (d *StreamRange) Next() (*Datapoint, error) {
	dp, err := d.dr.Next()

	//If there is an explicit error - or if there was a datapoint returned, just go with it
	if err != nil || dp != nil {
		d.index++
		return dp, err
	}

	return nil, nil
}
