package interpolators

import (
	"connectordb/streamdb/dataset/pipeline"
	"connectordb/streamdb/datastream"
	"errors"
)

//DefaultInterpolator is the one used when no interpolator is specified
var DefaultInterpolator = "closest"

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

//GetInterpolator gets an interpolator given the string which defines the interpolator and all arguments
//it takes
func GetInterpolator(dr datastream.DataRange, interp string) (Interpolator, error) {
	if interp == "" {
		interp = DefaultInterpolator
	}
	p, err := pipeline.ParsePipeline(interp)
	if err != nil {
		return nil, err
	}
	if len(p) != 1 {
		return nil, errors.New("There must be exactly one interpolator defined")
	}
	ifunc, ok := Interpolators[p[0].Symbol]
	if !ok {
		return nil, errors.New("Could not find '" + p[0].Symbol + "' interpolator.")
	}
	return ifunc(dr, p[0].Args)

}
