package query

import (
	"connectordb/streamdb/datastream"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	dpa1 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1},
		datastream.Datapoint{Timestamp: 2},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 4},
		datastream.Datapoint{Timestamp: 5},
	}

	dpa2 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1.1},
		datastream.Datapoint{Timestamp: 1.2},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3.2},
		datastream.Datapoint{Timestamp: 5.1},
		datastream.Datapoint{Timestamp: 5.2},
	}

	dpa := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1},
		datastream.Datapoint{Timestamp: 1.1},
		datastream.Datapoint{Timestamp: 1.2},
		datastream.Datapoint{Timestamp: 2},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3},
		datastream.Datapoint{Timestamp: 3.2},
		datastream.Datapoint{Timestamp: 4},
		datastream.Datapoint{Timestamp: 5},
		datastream.Datapoint{Timestamp: 5.1},
		datastream.Datapoint{Timestamp: 5.2},
	}
	mq := NewMockOperator(map[string]datastream.DatapointArray{"u/d/s1": dpa1, "u/d/s2": dpa2})

	dr, err := Merge(mq, []*StreamQuery{
		&StreamQuery{Stream: "u/d/s1", T1: 0.5},
		&StreamQuery{Stream: "u/d/s2"},
	})

	require.NoError(t, err)
	CompareRange(t, dr, dpa)

	_, err = Merge(mq, []*StreamQuery{
		&StreamQuery{Stream: "u/d/s2"},
		&StreamQuery{T1: 1, I1: 1},
	})
	require.Error(t, err)

	_, err = Merge(mq, []*StreamQuery{
		&StreamQuery{Stream: "u/d/s1"},
		&StreamQuery{Stream: "u/d/dne"},
	})
	require.Error(t, err)
}
