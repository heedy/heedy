package datastream

//The DataRange interface - this is the object that is returned from different caches/stores - it represents
//a range of data values stored in a certain way, and Next() gets the next datapoint in the range.
type DataRange interface {
	NextArray() (*DatapointArray, error) //Returns the next chunk of datapoints from the DataRange
	Next() (*Datapoint, error)           //Returns the next datapoint in the sequence
	Close()
}

//The EmptyRange is a range that always returns nil - as if there were no datapoints left.
//It is the DataRange equivalent of nil
type EmptyRange struct{}

//Close does absolutely nothing
func (r EmptyRange) Close() {}

//NextArray always just returns nil,nil
func (r EmptyRange) NextArray() (*DatapointArray, error) {
	return nil, nil
}

//Next always just returns nil,nil
func (r EmptyRange) Next() (*Datapoint, error) {
	return nil, nil
}

//A TimeRange is a Datarange which is time-bounded from both sides. That is, the datapoints allowed are only
//within the given time range. So if given a DataRange with range [a,b], and the timerange is (c,d], the
//TimeRange will return all datapoints within the Datarange which are within (c,d].
type TimeRange struct {
	dr        DataRange //The DataRange to wrap
	starttime float64   //The time at which to start the time range
	endtime   float64   //The time at which to stop returning datapoints
}

//Close closes the internal DataRange
func (r *TimeRange) Close() {
	r.dr.Close()
}

//NextArray returns the next datapoint array in sequence from the underlying DataRange, so long as it is within the
//correct timestamp bounds
func (r *TimeRange) NextArray() (*DatapointArray, error) {
	dpap, err := r.dr.NextArray()

	if err != nil || dpap == nil {
		return dpap, err
	}

	dpa := dpap.TRange(r.starttime, r.endtime)
	if dpa == nil {
		return nil, err
	}
	if dpa.Length() > 0 {
		return &dpa, err
	}

	return r.NextArray()
}

//Next returns the next datapoint in sequence from the underlying DataRange, so long as it is within the
//correct timestamp bounds
func (r *TimeRange) Next() (*Datapoint, error) {
	dp, err := r.dr.Next()
	//Skip datapoints before the starttime
	for dp != nil && dp.Timestamp <= r.starttime {
		dp, err = r.dr.Next()
	}
	//Return nil if the timestamp is beyond our range
	if dp != nil && r.endtime > 0 && dp.Timestamp > r.endtime {
		//The datapoint is beyond our range.
		return nil, nil
	}
	return dp, err
}

//NewTimeRange creates a time range given the time range of valid datapoints
func NewTimeRange(dr DataRange, starttime float64, endtime float64) *TimeRange {
	return &TimeRange{dr, starttime, endtime}
}

//NumRange returns only the first given number of datapoints (with an optional skip param) from a DataRange
type NumRange struct {
	dr      DataRange
	numleft int64 //The number of datapoints left to return
}

//Close closes the internal DataRange
func (r *NumRange) Close() {
	r.dr.Close()
}

//NextArray returns the next datapoint from the underlying DataRange so long as the datapoint array is within the
//amount of datapoints to return.
func (r *NumRange) NextArray() (*DatapointArray, error) {
	if r.numleft == 0 {
		return nil, nil
	}

	dpa, err := r.dr.NextArray()
	if err != nil {
		return nil, err
	}

	if int64(dpa.Length()) <= r.numleft {
		r.numleft -= int64(dpa.Length())
		return dpa, nil
	}
	dpa = dpa.IRange(0, int(r.numleft))
	r.numleft = 0
	return dpa, nil
}

//Next returns the next datapoint from the underlying DataRange so long as the datapoint is within the
//amonut of datapoints to return.
func (r *NumRange) Next() (*Datapoint, error) {
	if r.numleft == 0 {
		return nil, nil
	}
	r.numleft--
	return r.dr.Next()
}

//Skip the given number of datapoints without changing the number of datapoints left to return
func (r *NumRange) Skip(num int) error {
	for i := 0; i < num; i++ {
		_, err := r.dr.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

//NewNumRange initializes a new NumRange which will return up to the given amount of datapoints.
func NewNumRange(dr DataRange, datapoints int64) *NumRange {
	return &NumRange{dr, datapoints}
}

//DatapointArrayRange allows DatapointArray to conform to the range interface
type DatapointArrayRange struct {
	rangeindex int
	da         DatapointArray
}

//Close resets the range
func (d *DatapointArrayRange) Close() {
	d.rangeindex = 0
}

//Next returns the next datapoint
func (d *DatapointArrayRange) Next() (*Datapoint, error) {
	if d.rangeindex >= d.da.Length() {
		return nil, nil
	}
	d.rangeindex++
	return &d.da[d.rangeindex-1], nil
}

//NextArray returns what is left of the array
func (d *DatapointArrayRange) NextArray() (*DatapointArray, error) {
	if d.rangeindex >= d.da.Length() {
		return nil, nil
	}
	dpa := d.da[d.rangeindex:]
	d.rangeindex = d.da.Length()
	return &dpa, nil
}

//NewDatapointArrayRange does exactly what the function says
func NewDatapointArrayRange(da DatapointArray) *DatapointArrayRange {
	return &DatapointArrayRange{0, da}
}
