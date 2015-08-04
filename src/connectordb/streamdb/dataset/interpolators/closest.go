package interpolators

import (
	"connectordb/streamdb/datastream"
	"errors"
	"math"
)

//ClosestInterpolator interpolates a datarange by timestamp - getting the datapoint with the closest timestamp
type ClosestInterpolator struct {
	prevDatapoint *datastream.Datapoint
	curDatapoint  *datastream.Datapoint

	currentRange datastream.DataRange
}

//Next gets the datapoint corresponding to the interpolation timestamp
func (i *ClosestInterpolator) Interpolate(ts float64) (dp *datastream.Datapoint, err error) {

	for i.curDatapoint != nil && i.curDatapoint.Timestamp <= ts {
		i.prevDatapoint = i.curDatapoint
		i.curDatapoint, err = i.currentRange.Next()
		if err != nil {
			return nil, err
		}
	}
	if i.prevDatapoint == nil {
		return i.curDatapoint, nil
	}
	if i.curDatapoint == nil {
		return i.prevDatapoint, nil
	}
	//Both prev and cur are not nil. Find which one is closer to ts
	if math.Abs(i.prevDatapoint.Timestamp-ts) <= math.Abs(i.curDatapoint.Timestamp-ts) {
		return i.prevDatapoint, nil
	}
	return i.curDatapoint, nil
}

//Close the interpolator
func (i *ClosestInterpolator) Close() {
	i.currentRange.Close()
}

//NewClosestInterpolator returns the ClosestInterpolator for the given stream and starting time
func NewClosestInterpolator(dr datastream.DataRange, args []string) (Interpolator, error) {
	if len(args) > 0 {
		return nil, errors.New("before interpolator does not accept arguments")
	}

	pd, err := dr.Next()
	if err != nil {
		return nil, err
	}
	cd, err := dr.Next()

	return &ClosestInterpolator{pd, cd, dr}, err
}
