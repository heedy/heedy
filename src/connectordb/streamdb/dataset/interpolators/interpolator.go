package interpolators

import "connectordb/streamdb/datastream"

//Interpolator is an interface which given a timestamp, returns the appropriate
//datapoint. Interpolator is guaranteed to be called with increasing or equal timestamps
//since the dataset is constructed iteratively. "registered" interpolators are given
//a DataRange and a string array of arguments passed by the user
type Interpolator interface {
	Interpolate(timestamp float64) (tdp *datastream.Datapoint, err error)
	Close()
}

//InterpolatorGenerator is the signature of a function that generates an interpolator
type InterpolatorGenerator func(dr datastream.DataRange, args []string) (Interpolator, error)

//Interpolators is the map of all registered interpolations
var Interpolators = map[string]InterpolatorGenerator{
	"before":  NewBeforeInterpolator,
	"after":   NewAfterInterpolator,
	"closest": NewClosestInterpolator,
}
