/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package transforms

import (
	"strings"

	"github.com/connectordb/duck"
)

// The identity function, returns whatever was passed in.
func ConstantValueGenerator(value interface{}, inputError error) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if te.Flag == constantCheck {
			te.Flag = constantCheckTrue
		}

		return te.SetError(inputError).SetData(value)
	}
}

func PipelineGeneratorIdentity() TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		return te
	}
}

func pipelineGeneratorGet(propertyNames []string) TransformFunc {
	interfaceProps := make([]interface{}, len(propertyNames))
	for i, v := range propertyNames {
		interfaceProps[i] = v
	}

	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		data, ok := duck.Get(te.Datapoint.Data, interfaceProps...)

		if !ok {
			errStr := strings.Join(propertyNames, ", ")
			return te.Copy().SetErrorString("Could not find element [" + errStr + "] in " + duck.JSONString(te.Datapoint))
		}

		return te.SetData(data)
	}
}

func pipelineGeneratorSet(propertyNames []string, value TransformFunc) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		valueSide := te.Copy().Apply(value)
		if !valueSide.CanProcess() {
			return valueSide
		}

		if len(propertyNames) == 0 {
			return te.SetData(valueSide.Datapoint.Data)
		}

		return te.SetErrorString("Don't know how to set children yet!")
	}
}

func pipelineGeneratorHas(propertyName string) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		_, ok := duck.Get(te.Datapoint.Data, propertyName)
		return te.SetData(ok)
	}
}
