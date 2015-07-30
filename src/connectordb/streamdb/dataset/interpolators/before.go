package interpolators

import . "connectordb/streamdb/datastream"

//BeforeInterpolator interpolates a datarange by timestamp - getting the first
//datapoint BEFORE the given time.
//NOTE: This is a very, very preliminary interpolator. It is not clever at all in
//	the way it gets datapoints, in particular, it doesn't have any way to fix mismatched size
//	streams (ie, what if there are 1 million datapoints but there are 2 original timestamps,
//	meaning that this interpolator would iterate through ALL 1 million to get to the 2 that it cares
//	about, certainly not the best way to go about things)
type BeforeInterpolator struct {
	prevDatapoint *Datapoint
	curDatapoint  *Datapoint

	currentRange DataRange
}

//Next gets the datapoint corresponding to the interpolation timestamp
func (i *BeforeInterpolator) Next(ts float64) (dp *Datapoint, err error) {

	for i.curDatapoint != nil && i.curDatapoint.Timestamp <= ts {
		i.prevDatapoint = i.curDatapoint
		i.curDatapoint, err = i.currentRange.Next()
		if err != nil {
			return nil, err
		}
	}
	if i.prevDatapoint != nil && i.prevDatapoint.Timestamp > ts {
		return nil, nil
	}
	return i.prevDatapoint, nil
}

//Close the interpolator
func (i *BeforeInterpolator) Close() {
	i.currentRange.Close()
}

//NewBeforeInterpolator returns the BeforeInterpolator for the given stream and starting time
func NewBeforeInterpolator(ds *DataStream, device, stream int64, substream string, starttime float64) (*BeforeInterpolator, error) {
	dr, err := ds.TimePlusIndexRange(device, stream, substream, starttime, -1)
	if err != nil {
		return nil, err
	}
	pd, err := dr.Next()
	if err != nil {
		return nil, err
	}
	cd, err := dr.Next()

	return &BeforeInterpolator{pd, cd, dr}, err
}
