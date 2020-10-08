package timeseries

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTDataset(t *testing.T) {
	adb, oid1, oid2, cleanup := newDBWithObjects(t)
	defer cleanup()
	dpa1 := NewDatapointArrayIterator(DatapointArray{
		&Datapoint{Timestamp: 1, Data: 1},
		&Datapoint{Timestamp: 2, Data: 2},
		&Datapoint{Timestamp: 3, Data: 5},
		&Datapoint{Timestamp: 4, Data: 6},
		&Datapoint{Timestamp: 5, Data: 7},
	})

	dpa2 := NewDatapointArrayIterator(DatapointArray{
		&Datapoint{Timestamp: 1.1, Data: 1},
		&Datapoint{Timestamp: 2.1, Data: 2},
		&Datapoint{Timestamp: 2.9, Data: 3},
		&Datapoint{Timestamp: 3.5, Data: 4},
		&Datapoint{Timestamp: 3.9, Data: 5},
	})
	sd := TimeseriesDB{DB: adb,
		BatchSize:             3,
		MaxBatchSize:          5,
		BatchCompressionLevel: 3}
	TSDB = sd // need to set the global
	err := sd.Insert(oid1, dpa1, &InsertQuery{})
	require.NoError(t, err)
	err = sd.Insert(oid2, dpa2, &InsertQuery{})
	require.NoError(t, err)

	di, err := (&Dataset{
		Query: Query{
			T1: 0.0,
			T2: 5.0,
		},
		Dt: 1.0,
		Dataset: map[string]*DatasetElement{
			"x": &DatasetElement{
				Merge: []*Query{
					&Query{
						Timeseries: oid1,
					},
					&Query{
						Timeseries: oid2,
					},
				},
			},
		},
	}).Get(adb)
	require.NoError(t, err)

	dpa, err := NewArrayFromIterator(&TransformIterator{dpi: di, it: di})
	require.NoError(t, err)

	result := DatapointArray{
		&Datapoint{Timestamp: 0, Data: map[string]interface{}{
			"x": 1,
		}},
		&Datapoint{Timestamp: 1, Data: map[string]interface{}{
			"x": 1,
		}},
		&Datapoint{Timestamp: 2, Data: map[string]interface{}{
			"x": 2,
		}},
		&Datapoint{Timestamp: 3, Data: map[string]interface{}{
			"x": 5,
		}},
		&Datapoint{Timestamp: 4, Data: map[string]interface{}{
			"x": 6,
		}},
	}

	require.Equal(t, dpa.String(), result.String())

}

func TestXDataset(t *testing.T) {
	adb, oid1, oid2, cleanup := newDBWithObjects(t)
	defer cleanup()

	dpa1 := NewDatapointArrayIterator(DatapointArray{
		&Datapoint{Timestamp: 1, Data: 1},
		&Datapoint{Timestamp: 2, Data: 2},
		&Datapoint{Timestamp: 3, Data: 3},
		&Datapoint{Timestamp: 3.01, Data: 4},
		&Datapoint{Timestamp: 3.02, Data: 5},
		&Datapoint{Timestamp: 4, Data: 6},
		&Datapoint{Timestamp: 5, Data: 7},
	})

	dpa2 := NewDatapointArrayIterator(DatapointArray{
		&Datapoint{Timestamp: 1.1, Data: 1},
		&Datapoint{Timestamp: 2.1, Data: 2},
		&Datapoint{Timestamp: 2.9, Data: 3},
		&Datapoint{Timestamp: 3.5, Data: 4},
		&Datapoint{Timestamp: 3.9, Data: 5},
	})
	sd := TimeseriesDB{DB: adb,
		BatchSize:             3,
		MaxBatchSize:          6,
		BatchCompressionLevel: 3}
	TSDB = sd // need to set the global

	err := sd.Insert(oid1, dpa1, &InsertQuery{})
	require.NoError(t, err)
	err = sd.Insert(oid2, dpa2, &InsertQuery{})
	require.NoError(t, err)

	di, err := (&Dataset{
		Query: Query{
			Timeseries: oid1,
		},
		Dataset: map[string]*DatasetElement{
			"y": &DatasetElement{
				Query: Query{
					Timeseries: oid2,
				},
			},
		},
	}).Get(adb)
	require.NoError(t, err)

	dpa, err := NewArrayFromIterator(&TransformIterator{dpi: di, it: di})
	require.NoError(t, err)

	result := DatapointArray{
		&Datapoint{Timestamp: 1, Data: map[string]interface{}{
			"y": 1,
			"x": 1,
		}},
		&Datapoint{Timestamp: 2, Data: map[string]interface{}{
			"y": 2,
			"x": 2,
		}},
		&Datapoint{Timestamp: 3, Data: map[string]interface{}{
			"y": 3,
			"x": 3,
		}},
		&Datapoint{Timestamp: 3.01, Data: map[string]interface{}{
			"y": 3,
			"x": 4,
		}},
		&Datapoint{Timestamp: 3.02, Data: map[string]interface{}{
			"y": 3,
			"x": 5,
		}},
		&Datapoint{Timestamp: 4, Data: map[string]interface{}{
			"y": 5,
			"x": 6,
		}},
		&Datapoint{Timestamp: 5, Data: map[string]interface{}{
			"y": 5,
			"x": 7,
		}},
	}
	require.Equal(t, dpa.String(), result.String())

	di, err = (&Dataset{
		Query: Query{
			Timeseries: oid1,
		},
		Dataset: map[string]*DatasetElement{
			"y": &DatasetElement{
				Query: Query{
					Timeseries: oid2,
				},
			},
		},
		PostTransform: "$('x')==$('y')",
	}).Get(adb)
	require.NoError(t, err)

	dpa, err = NewArrayFromIterator(&TransformIterator{dpi: di, it: di})
	require.NoError(t, err)

	result = DatapointArray{
		&Datapoint{Timestamp: 1, Data: true},
		&Datapoint{Timestamp: 2, Data: true},
		&Datapoint{Timestamp: 3, Data: true},
		&Datapoint{Timestamp: 3.01, Data: false},
		&Datapoint{Timestamp: 3.02, Data: false},
		&Datapoint{Timestamp: 4, Data: false},
		&Datapoint{Timestamp: 5, Data: false},
	}

	require.Equal(t, dpa.String(), result.String())
}

func TestDatasetErrors(t *testing.T) {
	adb, oid1, oid2, cleanup := newDBWithObjects(t)
	defer cleanup()

	dpa1 := NewDatapointArrayIterator(DatapointArray{
		&Datapoint{Timestamp: 1, Data: 1},
		&Datapoint{Timestamp: 2, Data: 2},
		&Datapoint{Timestamp: 3, Data: 3},
		&Datapoint{Timestamp: 3.01, Data: 4},
		&Datapoint{Timestamp: 3.02, Data: 5},
		&Datapoint{Timestamp: 4, Data: 6},
		&Datapoint{Timestamp: 5, Data: 7},
	})

	dpa2 := NewDatapointArrayIterator(DatapointArray{
		&Datapoint{Timestamp: 1.1, Data: 1},
		&Datapoint{Timestamp: 2.1, Data: 2},
		&Datapoint{Timestamp: 2.9, Data: 3},
		&Datapoint{Timestamp: 3.5, Data: 4},
		&Datapoint{Timestamp: 3.9, Data: 5},
	})
	sd := TimeseriesDB{DB: adb,
		BatchSize:             3,
		MaxBatchSize:          5,
		BatchCompressionLevel: 3}
	TSDB = sd // need to set the global
	err := sd.Insert(oid1, dpa1, &InsertQuery{})
	require.NoError(t, err)
	err = sd.Insert(oid2, dpa2, &InsertQuery{})
	require.NoError(t, err)

	_, err = (&Dataset{
		Query: Query{
			Timeseries: oid1,
		},
		Dt: 1.0,
	}).Get(adb)
	require.Error(t, err)
	_, err = (&Dataset{
		Query: Query{
			Timeseries: oid1,
		},
		Dataset: map[string]*DatasetElement{
			"y": &DatasetElement{
				Query:        Query{Timeseries: oid2},
				Interpolator: "invalid",
			},
		},
	}).Get(adb)
	require.Error(t, err)
	_, err = (&Dataset{
		Query: Query{
			Timeseries: "blah/blah/blah",
		},
		Dataset: map[string]*DatasetElement{
			"y": &DatasetElement{
				Query:        Query{Timeseries: oid2},
				Interpolator: "closest",
			},
		},
	}).Get(adb)
	require.Error(t, err)
	_, err = (&Dataset{
		Query: Query{
			Timeseries: oid1,
		},
		Dataset: map[string]*DatasetElement{
			"y": &DatasetElement{
				Query:        Query{Timeseries: "dfsgfd"},
				Interpolator: "closest",
			},
		},
	}).Get(adb)
	require.Error(t, err)
	_, err = (&Dataset{
		Dataset: map[string]*DatasetElement{
			"y": &DatasetElement{
				Query:        Query{Timeseries: oid2},
				Interpolator: "closest",
			},
		},
	}).Get(adb)
	require.Error(t, err)
	_, err = (&Dataset{
		Dt: 1.3,
		Dataset: map[string]*DatasetElement{
			"y": &DatasetElement{
				Query:        Query{Timeseries: oid2},
				Interpolator: "closest",
			},
		},
	}).Get(adb)
	require.Error(t, err)

	_, err = (&Dataset{
		Query: Query{
			Timeseries: oid1,
		},
		Dataset: map[string]*DatasetElement{
			"x": &DatasetElement{
				Query:        Query{Timeseries: oid2},
				Interpolator: "closest",
			},
		},
	}).Get(adb)
	require.Error(t, err)

}
