package interpolators

import (
	"connectordb/datastream"
	"errors"
)

//BeforeInterpolator interpolates a datarange by timestamp - getting the first
//datapoint BEFORE the given time.
//NOTE: This is a very, very preliminary interpolator. It is not clever at all in
//	the way it gets datapoints, in particular, it doesn't have any way to fix mismatched size
//	streams (ie, what if there are 1 million datapoints but there are 2 original timestamps,
//	meaning that this interpolator would iterate through ALL 1 million to get to the 2 that it cares
//	about, certainly not the best way to go about things)
type BeforeInterpolator struct {
	prevDatapoint *datastream.Datapoint
	curDatapoint  *datastream.Datapoint

	currentRange datastream.DataRange
}

//Interpolate gets the datapoint corresponding to the interpolation timestamp
func (i *BeforeInterpolator) Interpolate(ts float64) (dp *datastream.Datapoint, err error) {

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

var before = InterpolatorDescription{
	Name:        "before",
	Description: "Returns the closest datapoint with a timestamp smaller than the dataset reference",

	Generator: func(dr datastream.DataRange, args []string) (Interpolator, error) {
		if len(args) > 0 {
			return nil, errors.New("before interpolator does not accept arguments")
		}
		pd, err := dr.Next()
		if err != nil {
			return nil, err
		}
		cd, err := dr.Next()

		return &BeforeInterpolator{pd, cd, dr}, err
	},
}
