package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
)

// The map that holds all custom functions
var (
	customFunctionMap = map[string]CustomFunctionGenerator{}
)

type CustomFunctionGenerator func(FunctionName string, Children ...TransformFunc) (TransformFunc, error)

// Regesters a function lambda under the name alias so it can be called in the
// pipeline. Note that the alias can only have a-zA-Z0-9 and underline, it also
// cannot start with a digit (standard function naming stuff for most langs.)
func RegisterCustomFunction(name string, function CustomFunctionGenerator) {
	customFunctionMap[name] = function
}

// Creates a function that always errs
func createInvalidFuncion(funcname string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		return nil, errors.New("Invalid Function:" + funcname)
	}
}

// Gets a custom function by name.
func getCustomFunction(identifier string, children ...TransformFunc) (TransformFunc, error) {
	function, ok := customFunctionMap[identifier]

	if !ok {
		return PipelineGeneratorIdentity(), errors.New("Undefined Function:" + identifier)
	}

	transformFunction, err := function(identifier, children...)

	if err != nil {
		return PipelineGeneratorIdentity(), errors.New("Undefined Function:" + identifier)
	}

	return transformFunction, nil
}
