/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package datastream

//The ExtendedDataRange interface - this is the object that is returned from different caches/stores - it represents
//a range of data values stored in a certain way, and Next() gets the next datapoint in the range.
type ExtendedDataRange interface {
	Index() int64                        //Returns the index of the ExtendedDataRange's next datapoint
	NextArray() (*DatapointArray, error) //Returns the next chunk of datapoints from the ExtendedDataRange
	Next() (*Datapoint, error)           //Returns the next datapoint in the sequence
	Close()
}

//DataRange is ExtendedDataRange's little brother - while ExtendedDataRange contains the NextArray and Index methods for more powerful
//manipulation, a DataRange is just a basic iterator. Note that all ExtendedDataRange fit the  DataRange interface
type DataRange interface {
	Next() (*Datapoint, error)
	Close()
}

//The EmptyRange is a range that always returns nil - as if there were no datapoints left.
//It is the ExtendedDataRange equivalent of nil
type EmptyRange struct{}

//Index just returns 0
func (r EmptyRange) Index() int64 {
	return 0
}

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
//within the given time range. So if given a ExtendedDataRange with range [a,b], and the timerange is (c,d], the
//TimeRange will return all datapoints within the Datarange which are within (c,d].
type TimeRange struct {
	dr      ExtendedDataRange //The ExtendedDataRange to wrap
	endtime float64         //The time at which to stop returning datapoints
	dpap    *DatapointArray //The current array that is being read
}

//Index returns the underlying ExtendedDataRange's index.
func (r *TimeRange) Index() int64 {
	if r.dpap != nil {
		return r.dr.Index() - int64(r.dpap.Length())
	}
	return r.dr.Index()
}

//Close closes the internal ExtendedDataRange
func (r *TimeRange) Close() {
	r.dr.Close()
}

//NextArray returns the next datapoint array in sequence from the underlying ExtendedDataRange, so long as it is within the
//correct timestamp bounds
func (r *TimeRange) NextArray() (dpap *DatapointArray, err error) {
	if r.dpap == nil {
		r.dpap, err = r.dr.NextArray()
	}

	if err != nil || r.dpap == nil {
		return r.dpap, err
	}

	dpa := r.dpap.TEnd(r.endtime)
	r.dpap = nil
	if dpa == nil {
		return nil, err
	}
	if dpa.Length() > 0 {
		return &dpa, err
	}

	return nil, nil
}

//Next returns the next datapoint in sequence from the underlying ExtendedDataRange, so long as it is within the
//correct timestamp bounds
func (r *TimeRange) Next() (dp *Datapoint, err error) {
	if r.dpap != nil && r.dpap.Length() > 0 {
		res := (*r.dpap)[0]
		dpa := (*r.dpap)[1:]
		r.dpap = &dpa
		if r.dpap.Length() == 0 {
			r.dpap = nil
		}
		dp = &res
	} else {
		dp, err = r.dr.Next()
	}

	//Return nil if the timestamp is beyond our range
	if dp != nil && r.endtime > 0.0 && dp.Timestamp > r.endtime {
		//The datapoint is beyond our range.
		return nil, nil
	}
	return dp, err
}

//NewTimeRange creates a time range given the time range of valid datapoints
func NewTimeRange(dr ExtendedDataRange, starttime float64, endtime float64) (ExtendedDataRange, error) {

	//We have a ExtendedDataRange - but we don't know what time it starts at. We want to skip the
	// datapoints before starttime
	dpap, err := dr.NextArray()
	for dpap != nil && err == nil {
		dpa := dpap.TStart(starttime)
		if dpa.Length() > 0 {
			return &TimeRange{dr, endtime, &dpa}, nil
		}
		dpap, err = dr.NextArray()
	}
	return EmptyRange{}, err
}

//NumRange returns only the first given number of datapoints (with an optional skip param) from a ExtendedDataRange
type NumRange struct {
	dr      ExtendedDataRange
	numleft int64 //The number of datapoints left to return
}

//Close closes the internal ExtendedDataRange
func (r *NumRange) Close() {
	r.dr.Close()
}

//Index returns the underlying ExtendedDataRange's index value
func (r *NumRange) Index() int64 {
	return r.dr.Index()
}

//NextArray returns the next datapoint from the underlying ExtendedDataRange so long as the datapoint array is within the
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

//Next returns the next datapoint from the underlying ExtendedDataRange so long as the datapoint is within the
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
func NewNumRange(dr ExtendedDataRange, datapoints int64) *NumRange {
	return &NumRange{dr, datapoints}
}

//DatapointArrayRange allows DatapointArray to conform to the range interface
type DatapointArrayRange struct {
	rangeindex int
	da         DatapointArray
	startindex int64
}

//Close resets the range
func (d *DatapointArrayRange) Close() {
	d.rangeindex = 0
}

//Index returns the index of the DatapointArray
func (d *DatapointArrayRange) Index() int64 {
	return d.startindex + int64(d.rangeindex)
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
func NewDatapointArrayRange(da DatapointArray, startindex int64) *DatapointArrayRange {
	return &DatapointArrayRange{0, da, startindex}
}
