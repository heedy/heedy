package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
	"strings"

	"github.com/connectordb/duck"
)

// The identity function, returns whatever was passed in.
func pipelineGeneratorConstant(value interface{}, inputError error) TransformFunc {
	dpp := &datastream.Datapoint{
		Data: value,
	}
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		return dpp, inputError
	}
}

func PipelineGeneratorIdentity() TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		return dp, nil
	}
}

func pipelineGeneratorGet(propertyNames []string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		interfaceProps := make([]interface{}, len(propertyNames))
		for i, v := range propertyNames {
			interfaceProps[i] = v
		}

		var ok bool
		result := CopyDatapoint(dp)
		result.Data, ok = duck.Get(dp.Data, interfaceProps...)

		if !ok {
			errStr := strings.Join(propertyNames, ", ")
			return nil, errors.New("Could not find element [" + errStr + "] in " + duck.JSONString(dp))
		}

		return result, nil
	}
}

func pipelineGeneratorSet(propertyNames []string, value TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		rightHandSide, err := value(dp)
		if err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)

		if len(propertyNames) == 0 {
			result.Data = rightHandSide.Data
			return result, nil
		}

		return nil, errors.New("Don't know how to set children yet!")
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
