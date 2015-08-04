package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"

	"github.com/connectordb/duck"
)

// The identity function, returns whatever was passed in.
func pipelineGeneratorConstant(value interface{}, inputError error) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		result := CopyDatapoint(dp)
		result.Data = value
		return result, inputError
	}
}

func pipelineGeneratorIdentity() TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		return dp, nil
	}
}

func pipelineGeneratorGet(propertyName string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		var ok bool
		result := CopyDatapoint(dp)
		result.Data, ok = duck.Get(dp.Data, propertyName)

		if !ok {
			return nil, errors.New("Could not find element '" + propertyName + "' in " + duck.JSONString(dp))
		}

		return result, nil
	}
}

func pipelineGeneratorHas(propertyName string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}
		result := CopyDatapoint(dp)
		_, result.Data = duck.Get(dp.Data, propertyName)
		return result, nil

	}
}
