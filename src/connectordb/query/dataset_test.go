/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package query

import (
	"connectordb/datastream"
	"testing"

	"github.com/connectordb/pipescript"
	"github.com/stretchr/testify/require"
)

func TestDatasetErrors(t *testing.T) {
	dpa1 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1, Data: 1},
		datastream.Datapoint{Timestamp: 2, Data: 2},
		datastream.Datapoint{Timestamp: 3, Data: 3},
		datastream.Datapoint{Timestamp: 3, Data: 4},
		datastream.Datapoint{Timestamp: 3, Data: 5},
		datastream.Datapoint{Timestamp: 4, Data: 6},
		datastream.Datapoint{Timestamp: 5, Data: 7},
	}

	dpa2 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1.1, Data: 1},
		datastream.Datapoint{Timestamp: 2.1, Data: 2},
		datastream.Datapoint{Timestamp: 2.9, Data: 3},
		datastream.Datapoint{Timestamp: 3.5, Data: 4},
		datastream.Datapoint{Timestamp: 3.9, Data: 5},
	}

	mq := NewMockOperator(map[string]datastream.DatapointArray{"a/b/c": dpa1, "d/e/f": dpa2})

	_, err := DatasetQuery{
		StreamQuery: StreamQuery{
			Stream: "a/b/c",
		},
		Dt: 1.0,
	}.Run(mq)
	require.Error(t, err)
	_, err = DatasetQuery{
		StreamQuery: StreamQuery{
			Stream: "a/b/c",
		},
	}.Run(mq)
	require.Error(t, err)
	_, err = DatasetQuery{
		StreamQuery: StreamQuery{
			Stream: "a/b/c",
		},
		Dataset: map[string]*DatasetQueryElement{
			"y": &DatasetQueryElement{
				StreamQuery:  StreamQuery{Stream: "d/e/f"},
				Interpolator: "invalid",
			},
		},
	}.Run(mq)
	require.Error(t, err)
	_, err = DatasetQuery{
		StreamQuery: StreamQuery{
			Stream: "blah/blah/blah",
		},
		Dataset: map[string]*DatasetQueryElement{
			"y": &DatasetQueryElement{
				StreamQuery:  StreamQuery{Stream: "d/e/f"},
				Interpolator: "closest",
			},
		},
	}.Run(mq)
	require.Error(t, err)
	_, err = DatasetQuery{
		StreamQuery: StreamQuery{
			Stream: "a/b/c",
		},
		Dataset: map[string]*DatasetQueryElement{
			"y": &DatasetQueryElement{
				StreamQuery:  StreamQuery{Stream: "blah/blah/blah"},
				Interpolator: "closest",
			},
		},
	}.Run(mq)
	require.Error(t, err)
	_, err = DatasetQuery{
		Dataset: map[string]*DatasetQueryElement{
			"y": &DatasetQueryElement{
				StreamQuery:  StreamQuery{Stream: "d/e/f"},
				Interpolator: "closest",
			},
		},
	}.Run(mq)
	require.Error(t, err)
	_, err = DatasetQuery{
		Dt: 1.3,
		Dataset: map[string]*DatasetQueryElement{
			"y": &DatasetQueryElement{
				StreamQuery:  StreamQuery{Stream: "d/e/f"},
				Interpolator: "closest",
			},
		},
	}.Run(mq)
	require.Error(t, err)

	_, err = DatasetQuery{
		StreamQuery: StreamQuery{
			Stream: "a/b/c",
		},
		Dataset: map[string]*DatasetQueryElement{
			"x": &DatasetQueryElement{
				StreamQuery:  StreamQuery{Stream: "d/e/f"},
				Interpolator: "closest",
			},
		},
	}.Run(mq)
	require.Error(t, err)

}

func TestYDatasetBasics(t *testing.T) {
	dpa1 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1, Data: 1},
		datastream.Datapoint{Timestamp: 2, Data: 2},
		datastream.Datapoint{Timestamp: 3, Data: 3},
		datastream.Datapoint{Timestamp: 3, Data: 4},
		datastream.Datapoint{Timestamp: 3, Data: 5},
		datastream.Datapoint{Timestamp: 4, Data: 6},
		datastream.Datapoint{Timestamp: 5, Data: 7},
	}

	dpa2 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1.1, Data: 1},
		datastream.Datapoint{Timestamp: 2.1, Data: 2},
		datastream.Datapoint{Timestamp: 2.9, Data: 3},
		datastream.Datapoint{Timestamp: 3.5, Data: 4},
		datastream.Datapoint{Timestamp: 3.9, Data: 5},
	}

	mq := NewMockOperator(map[string]datastream.DatapointArray{"a/b/c": dpa1, "d/e/f": dpa2})

	dr, err := DatasetQuery{
		StreamQuery: StreamQuery{
			Stream: "a/b/c",
		},
		Dataset: map[string]*DatasetQueryElement{
			"y": &DatasetQueryElement{
				StreamQuery:  StreamQuery{Stream: "d/e/f"},
				Interpolator: "closest",
			},
		},
	}.Run(mq)
	require.NoError(t, err)

	result := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1, Data: map[string]*datastream.Datapoint{
			"y": &datastream.Datapoint{Timestamp: 1.1, Data: 1},
			"x": &datastream.Datapoint{Timestamp: 1, Data: 1},
		}},
		datastream.Datapoint{Timestamp: 2, Data: map[string]*datastream.Datapoint{
			"y": &datastream.Datapoint{Timestamp: 2.1, Data: 2},
			"x": &datastream.Datapoint{Timestamp: 2, Data: 2},
		}},
		datastream.Datapoint{Timestamp: 3, Data: map[string]*datastream.Datapoint{
			"y": &datastream.Datapoint{Timestamp: 2.9, Data: 3},
			"x": &datastream.Datapoint{Timestamp: 3, Data: 3},
		}},
		datastream.Datapoint{Timestamp: 3, Data: map[string]*datastream.Datapoint{
			"y": &datastream.Datapoint{Timestamp: 2.9, Data: 3},
			"x": &datastream.Datapoint{Timestamp: 3, Data: 4},
		}},
		datastream.Datapoint{Timestamp: 3, Data: map[string]*datastream.Datapoint{
			"y": &datastream.Datapoint{Timestamp: 2.9, Data: 3},
			"x": &datastream.Datapoint{Timestamp: 3, Data: 5},
		}},
		datastream.Datapoint{Timestamp: 4, Data: map[string]*datastream.Datapoint{
			"y": &datastream.Datapoint{Timestamp: 3.9, Data: 5},
			"x": &datastream.Datapoint{Timestamp: 4, Data: 6},
		}},
		datastream.Datapoint{Timestamp: 5, Data: map[string]*datastream.Datapoint{
			"y": &datastream.Datapoint{Timestamp: 3.9, Data: 5},
			"x": &datastream.Datapoint{Timestamp: 5, Data: 7},
		}},
	}
	CompareRange(t, dr, result)
	dr.Close()
}

func TestTDatasetBasics(t *testing.T) {
	dpa1 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1, Data: 1},
		datastream.Datapoint{Timestamp: 2, Data: 2},
		datastream.Datapoint{Timestamp: 3, Data: 3},
		datastream.Datapoint{Timestamp: 3, Data: 4},
		datastream.Datapoint{Timestamp: 3, Data: 5},
		datastream.Datapoint{Timestamp: 4, Data: 6},
		datastream.Datapoint{Timestamp: 5, Data: 7},
	}

	dpa2 := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1.1, Data: 1},
		datastream.Datapoint{Timestamp: 2.1, Data: 2},
		datastream.Datapoint{Timestamp: 2.9, Data: 3},
		datastream.Datapoint{Timestamp: 3.5, Data: 4},
		datastream.Datapoint{Timestamp: 3.9, Data: 5},
	}

	mq := NewMockOperator(map[string]datastream.DatapointArray{"a/b/c": dpa1, "d/e/f": dpa2})

	dr, err := DatasetQuery{
		StreamQuery: StreamQuery{
			T2: 5,
		},
		Dt: 1.0,
		Dataset: map[string]*DatasetQueryElement{
			"x": &DatasetQueryElement{
				Merge: []*StreamQuery{
					&StreamQuery{
						Stream: "d/e/f",
					},
					&StreamQuery{
						Stream: "a/b/c",
					},
				},
				Interpolator: "closest",
			},
		},
	}.Run(mq)
	require.NoError(t, err)

	result := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 0, Data: map[string]*pipescript.Datapoint{
			"x": &pipescript.Datapoint{Timestamp: 1, Data: 1},
		}},
		datastream.Datapoint{Timestamp: 1, Data: map[string]*pipescript.Datapoint{
			"x": &pipescript.Datapoint{Timestamp: 1, Data: 1},
		}},
		datastream.Datapoint{Timestamp: 2, Data: map[string]*pipescript.Datapoint{
			"x": &pipescript.Datapoint{Timestamp: 2, Data: 2},
		}},
		datastream.Datapoint{Timestamp: 3, Data: map[string]*pipescript.Datapoint{
			"x": &pipescript.Datapoint{Timestamp: 3, Data: 5},
		}},
		datastream.Datapoint{Timestamp: 4, Data: map[string]*pipescript.Datapoint{
			"x": &pipescript.Datapoint{Timestamp: 4, Data: 6},
		}},
	}
	CompareRange(t, dr, result)
	dr.Close()
}
