package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterCustomFunction(t *testing.T) {
	testDatapoint := datastream.Datapoint{Data: true}

	// Regular func
	customGen := func(name string, children ...TransformFunc) (TransformFunc, error) {
		assert.Equal(t, "myname", name)
		return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
			return dp, nil
		}, nil
	}
	RegisterCustomFunction("myname", customGen)

	// function that should error out
	errGen := func(name string, children ...TransformFunc) (TransformFunc, error) {
		return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
			return dp, nil
		}, errors.New("foobar")
	}
	RegisterCustomFunction("errGen", errGen)

	childRequired := func(name string, children ...TransformFunc) (TransformFunc, error) {
		if len(children) != 1 {
			return pipelineGeneratorIdentity(), errors.New("requires exactly one child")
		}
		return pipelineGeneratorIdentity(), nil
	}
	RegisterCustomFunction("childRequired", childRequired)

	// regular func testing
	{
		custFunc, err := getCustomFunction("myname")
		require.Nil(t, err)
		result, err := custFunc(&testDatapoint)
		require.Nil(t, err)
		assert.Equal(t, &testDatapoint, result)
	}

	// invalid function testing
	{
		_, err := getCustomFunction("does not exist")
		require.NotNil(t, err)
	}

	// err function testing
	{
		_, err := getCustomFunction("errGen")
		require.NotNil(t, err)
	}

	// Child function testing
	{
		_, err := getCustomFunction("childRequired")
		require.NotNil(t, err)

		rightNumber, err := getCustomFunction("childRequired", pipelineGeneratorIdentity())
		require.Nil(t, err)

		result, err := rightNumber(&testDatapoint)
		require.Nil(t, err)
		assert.Equal(t, &testDatapoint, result)

	}

}
