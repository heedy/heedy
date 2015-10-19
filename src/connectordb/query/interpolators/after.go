package interpolators

import (
	"connectordb/datastream"
	"errors"
)

//AfterInterpolator interpolates a datarange by timestamp - getting the first
//datapoint AFTER the given time.
//NOTE: This is a very, very preliminary interpolator. It is not clever at all in
//	the way it gets datapoints, in particular, it doesn't have any way to fix mismatched size
//	streams (ie, what if there are 1 million datapoints but there are 2 original timestamps,
//	meaning that this interpolator would iterate through ALL 1 million to get to the 2 that it cares
//	about, certainly not the best way to go about things)
type AfterInterpolator struct {
	prevDatapoint *datastream.Datapoint

	currentRange datastream.DataRange
}

//Interpolate gets the datapoint corresponding to the interpolation timestamp
func (i *AfterInterpolator) Interpolate(ts float64) (dp *datastream.Datapoint, err error) {
	if i.prevDatapoint != nil && i.prevDatapoint.Timestamp > ts {
		return i.prevDatapoint, nil
	}

	//We no longer care about prevDatapoint - get a datapoint that satisfies the constraint...
	//or return nil
	//TODO: Use NextArray - it's faster
	i.prevDatapoint, err = i.currentRange.Next()
	for i.prevDatapoint != nil && err == nil && i.prevDatapoint.Timestamp <= ts {
		i.prevDatapoint, err = i.currentRange.Next()
	}
	return i.prevDatapoint, err
}

//Close the interpolator
func (i *AfterInterpolator) Close() {
	i.currentRange.Close()
}

var after = InterpolatorDescription{
	Name:        "after",
	Description: "Returns the closest datapoint with a timestamp greater than the dataset reference",

	Generator: func(dr datastream.DataRange, args []string) (Interpolator, error) {
		if len(args) > 0 {
			return nil, errors.New("after interpolator does not accept arguments")
		}
		return &AfterInterpolator{nil, dr}, nil
	},
}
