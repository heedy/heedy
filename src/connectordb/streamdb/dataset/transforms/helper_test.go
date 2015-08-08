package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleResultError(t *testing.T) {
	dp := datastream.Datapoint{}

	// has error defined
	assert.Error(t, handleResultError("prefix", &dp, errors.New("testerr"), true))

	// nil datapoint
	assert.Error(t, handleResultError("prefix", nil, nil, true))

	// bad coersion
	assert.Error(t, handleResultError("prefix", &dp, nil, false))

	// ok
	assert.Nil(t, handleResultError("prefix", &dp, nil, true))
}

func TestReadBool(t *testing.T) {
	dp := datastream.Datapoint{Data: true}

	{
		result, err := readBool("", &dp, PipelineGeneratorIdentity())
		assert.Nil(t, err)
		assert.True(t, result)
	}

	{
		_, err := readBool("", nil, PipelineGeneratorIdentity())
		assert.Error(t, err)
	}
}
