package datastream

//StreamRange is a DataRange that combines the redis and sql data into one coherent stream
type StreamRange struct {
	ds *DataStream
	dr DataRange

	index     int64
	deviceID  int64
	streamID  int64
	substream string
}

//Close the StreamRange
func (d *StreamRange) Close() {
	if d.dr != nil {
		d.dr.Close()
	}
}

//Index returns the current index of the values
func (d *StreamRange) Index() int64 {
	return d.index
}

func (d *StreamRange) getNextDataRange() (err error) {
	//If the program got here it means the datarange is empty.
	//This means we can sorta cheat. If the datarange is empty, it means that the sqlstore
	//ran out of data. This is because all of the data is in redis... UNLESS the batch
	//was just written right now.
	//If there was no batch written, IRange will return the datarange straight from redis.
	//If there WAS a batch written, IRange will return a StreamRange - which is also a DataRange.
	//Since writing batches in-between queries is something that rarely happens,
	//for simplicity's sake, we just stack the StreamRanges each time that happens.
	d.dr, err = d.ds.IRange(d.deviceID, d.streamID, d.substream, d.index, 0)
	return err
}

//NextArray returns the next datapoint array from the stream
func (d *StreamRange) NextArray() (*DatapointArray, error) {
	dpa, err := d.dr.NextArray()
	if err != nil || dpa != nil {
		d.index += int64(dpa.Length())
		return dpa, err
	}

	if err = d.getNextDataRange(); err != nil {
		return nil, err
	}

	dpa, err = d.dr.NextArray()
	if dpa != nil && err == nil {
		d.index += int64(dpa.Length())
	}
	return dpa, err
}

//Next returns the next datapoint
func (d *StreamRange) Next() (*Datapoint, error) {
	dp, err := d.dr.Next()

	//If there is an explicit error - or if there was a datapoint returned, just go with it
	if err != nil || dp != nil {
		d.index++
		return dp, err
	}

	if err = d.getNextDataRange(); err != nil {
		return nil, err
	}

	dp, err = d.dr.Next()
	if err == nil && dp != nil {
		d.index++
	}
	return dp, err
}
