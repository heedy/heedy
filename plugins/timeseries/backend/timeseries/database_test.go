package timeseries

import (
	"os"
	"testing"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/klauspost/compress/zstd"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	dpa1 = DatapointArray{&Datapoint{1.0, 0, "helloWorld", "me"}, &Datapoint{2.0, 0, "helloWorld2", "me2"}}
	dpa2 = DatapointArray{&Datapoint{1.0, 0, "helloWorl", "me"}, &Datapoint{2.0, 0, "helloWorld2", "me2"}}
	dpa3 = DatapointArray{&Datapoint{1.0, 0, "helloWorl", "me"}}

	dpa4 = DatapointArray{&Datapoint{3.0, 0, 12.0, ""}}

	//Warning: the map types change depending on marshaller/unmarshaller is used
	dpa5 = DatapointArray{&Datapoint{3.0, 0, map[string]interface{}{"hello": 2.0, "y": "hi"}, ""}}

	dpa6 = DatapointArray{&Datapoint{1.0, 0, 1.0, ""}, &Datapoint{2.0, 0, 2.0, ""}, &Datapoint{3.0, 0, 3., ""}, &Datapoint{4.0, 0, 4., ""}, &Datapoint{5.0, 0, 5., ""}}
	dpa7 = DatapointArray{
		&Datapoint{1., 1, "test0", ""},
		&Datapoint{2., .7, "test1", ""},
		&Datapoint{3., .6, "test2", ""},
		&Datapoint{4., .5, "test3", ""},
		&Datapoint{5., .4, "test4", ""},
		&Datapoint{6., 1, "test5", ""},
		&Datapoint{6., .2, "test6", ""},
		&Datapoint{7., .1, "test7", ""},
		&Datapoint{8., 0, "test8", ""},
	}
)

func TestArrayEquality(t *testing.T) {
	require.True(t, dpa1.IsEqual(dpa1))
	require.False(t, dpa1.IsEqual(dpa2))
	require.False(t, dpa2.IsEqual(dpa3))
	require.True(t, dpa4.IsEqual(dpa4))
	require.True(t, dpa5.IsEqual(dpa5))
}

func newAssets(t *testing.T) (*assets.Assets, func()) {
	a, err := assets.Open("", nil)
	require.NoError(t, err)
	os.RemoveAll("./test_db")
	a.FolderPath = "./test_db"
	sqla := "sqlite3://heedy.db?_journal=WAL&_fk=1"
	a.Config.SQL = &sqla

	assets.SetGlobal(a)
	return a, func() {
		os.RemoveAll("./test_db")
	}
}

func newDB(t *testing.T) (*database.AdminDB, func()) {
	a, cleanup := newAssets(t)
	a.Config.Verbose = true
	err := database.Create(a)
	if err != nil {
		cleanup()
	}
	require.NoError(t, err)

	db, err := database.Open(a)
	require.NoError(t, err)

	logrus.SetLevel(logrus.DebugLevel)

	zencoder, err = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevel(2)))
	require.NoError(t, err)

	return db, cleanup
}

func newDBWithUser(t *testing.T) (*database.AdminDB, func()) {
	adb, cleanup := newDB(t)

	name := "test"
	passwd := "test"
	require.NoError(t, adb.CreateUser(&database.User{
		UserName: &name,
		Password: &passwd,
	}))
	return adb, cleanup
}

func newDBWithObjects(t *testing.T) (*database.AdminDB, string, string, func()) {
	db, cleanup := newDBWithUser(t)
	oname := "myobject"
	otype := "timeseries"
	uname := "test"
	oid1, err := db.CreateObject(&database.Object{
		Details: database.Details{
			Name: &oname,
		},
		Type:  &otype,
		Owner: &uname,
	})
	require.NoError(t, err)
	oid2, err := db.CreateObject(&database.Object{
		Details: database.Details{
			Name: &oname,
		},
		Type:  &otype,
		Owner: &uname,
	})
	require.NoError(t, err)
	return db, oid1, oid2, cleanup
}

func cmpQuery(t *testing.T, s TimeseriesDB, q *Query, res DatapointArray) {
	di, err := s.Query(q)
	require.NoError(t, err)
	dpa, err := NewArrayFromIterator(di)
	require.NoError(t, err)
	require.True(t, res.IsEqual(dpa), "%s different from %s", res.String(), dpa.String())
}

func TestDatabase(t *testing.T) {
	adb, oid1, oid2, cleanup := newDBWithObjects(t)
	defer cleanup()

	s := TimeseriesDB{
		DB:                    adb,
		BatchSize:             3,
		MaxBatchSize:          5,
		BatchCompressionLevel: 2,
	}
	// TODO: SET ACTION FALSE - if ever reenable actions, set this to true
	action := false

	l, err := s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, l, int64(0))

	imethod := "insert"
	err = s.Insert(oid1, NewDatapointArrayIterator(dpa1), &InsertQuery{
		Actions: &action,
		Method:  &imethod,
	})
	require.NoError(t, err)

	di, err := s.Query(&Query{
		Timeseries: oid1,
		T:          "1.0",
		Actions:    &action,
	})
	require.NoError(t, err)
	dpa, err := NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 1)
	require.True(t, dpa.IsEqual(dpa1[0:1]), "%s different from %s", dpa1[0:1].String(), dpa.String())

	di, err = s.Query(&Query{
		Timeseries: oid1,
		T1:         "1.0",
		Actions:    &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 2)
	require.Equal(t, dpa.String(), dpa1.String())

	// Overwrite the first datapoint
	insertType := "update"
	err = s.Insert(oid1, NewDatapointArrayIterator(dpa3), &InsertQuery{
		Method:  &insertType,
		Actions: &action,
	})
	require.NoError(t, err)

	di, err = s.Query(&Query{
		Timeseries: oid1,
		T1:         "1.0",
		Actions:    &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 2)
	require.False(t, dpa.IsEqual(dpa1), "%s same as %s", dpa1.String(), dpa.String())
	require.True(t, dpa[0].IsEqual(dpa3[0]), "%s different from %s", dpa[0].String(), dpa3[0].String())
	require.True(t, dpa[1].IsEqual(dpa1[1]), "%s different from %s", dpa[1].String(), dpa1[1].String())

	l, err = s.Length(oid1, action)
	require.NoError(t, err)
	require.Equal(t, int64(2), l)

	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
		T:          "1.0",
		Actions:    &action,
	}))

	di, err = s.Query(&Query{
		Timeseries: oid1,
		T1:         1.0,
		Actions:    &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 1)
	require.True(t, dpa[0].IsEqual(dpa1[1]), "%s different from %s", dpa[0].String(), dpa1[1].String())

	err = s.Insert(oid2, NewDatapointArrayIterator(dpa7), &InsertQuery{
		Method: &imethod,
	})
	require.Error(t, err)

	itype := "append"
	err = s.Insert(oid2, NewDatapointArrayIterator(dpa7[:6]), nil)
	require.NoError(t, err)
	err = s.Insert(oid2, NewDatapointArrayIterator(dpa7[6:]), &InsertQuery{
		Method: &itype,
	})
	require.Error(t, err)
	err = s.Insert(oid2, NewDatapointArrayIterator(dpa7[6:]), nil)
	require.NoError(t, err)

	l, err = s.Length(oid2, false)
	require.NoError(t, err)
	require.Equal(t, int64(len(dpa7)-1), l) // dpa7 has timestamp 6 repeated

	i1 := int64(1)
	i2 := int64(-3)
	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I1:         &i1,
		I2:         &i2,
	}, dpa7[1:5])

	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I:          &i1,
	}, dpa7[1:2])

	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I:          &i2,
	}, dpa7[6:7]) // The t=6 element was replaced

	i3 := int64(3)
	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I1:         &i1,
		I2:         &i3,
	}, dpa7[1:3])
	i3 = -1
	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I1:         &i3,
	}, dpa7[len(dpa7)-1:])

	i2 = 80
	di, err = s.Query(&Query{
		Timeseries: oid2,
		I2:         &i2,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa7)-1, len(dpa))

	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I1:         &i1,
		I2:         &i3,
		T1:         float64(6),
		T2:         float64(40),
	}, dpa7[6:8])

	i8 := int64(8000)
	i9 := int64(-8000)
	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I1:         &i8,
	}, DatapointArray{})
	cmpQuery(t, s, &Query{
		Timeseries: oid2,
		I2:         &i9,
	}, DatapointArray{})

}

func TestEdgeCases(t *testing.T) {
	adb, oid1, _, cleanup := newDBWithObjects(t)
	defer cleanup()

	s := TimeseriesDB{
		DB:                    adb,
		BatchSize:             3,
		MaxBatchSize:          5,
		BatchCompressionLevel: 3,
	}

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
	}), nil))

	cmpQuery(t, s, &Query{
		Timeseries: oid1,
	}, DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
	})

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{3., 0, "test3", ""},
		&Datapoint{4., 0, "test4", ""},
		&Datapoint{5., 0, "test6", ""},
	}), nil))

	cmpQuery(t, s, &Query{
		Timeseries: oid1,
	}, DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
		&Datapoint{3., 0, "test3", ""},
		&Datapoint{4., 0, "test4", ""},
		&Datapoint{5., 0, "test6", ""},
	})

	// Clear the timeseries
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
	}))
	l, err := s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	// Now test the edge case where mergeBatch returns an empty array
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
		&Datapoint{3., 0, "test2", ""},

		&Datapoint{4., 0, "test3", ""},
		&Datapoint{5., 0, "test4", ""},
		&Datapoint{6., 0, "test5", ""},

		&Datapoint{7., 0, "test6", ""},
		&Datapoint{8., 0, "test6", ""},
		&Datapoint{9., 0, "test6", ""},

		&Datapoint{10., 0, "test6", ""},
		&Datapoint{11., 0, "test6", ""},
		&Datapoint{12., 0, "test6", ""},
	}), nil))

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{1., 5, "test0", ""},
		&Datapoint{10., 5, "test0", ""},
	}), nil))

	cmpQuery(t, s, &Query{
		Timeseries: oid1,
	}, DatapointArray{
		&Datapoint{1., 5, "test0", ""},
		&Datapoint{6., 0, "test5", ""},
		&Datapoint{7., 0, "test6", ""},
		&Datapoint{8., 0, "test6", ""},
		&Datapoint{9., 0, "test6", ""},
		&Datapoint{10., 5, "test0", ""},
	})

	// Clear the timeseries
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
	}))
	l, err = s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	// Now test where the insert is done after an appendUntil without going to a next batch
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
		&Datapoint{3., 0, "test2", ""},

		&Datapoint{4., 0, "test3", ""},
		&Datapoint{5., 0, "test4", ""},
		&Datapoint{6., 0, "test5", ""},

		&Datapoint{7., 0, "test6", ""},
		&Datapoint{8., 0, "test6", ""},
		&Datapoint{9., 0, "test6", ""},

		&Datapoint{10., 0, "test6", ""},
		&Datapoint{11., 0, "test6", ""},
		&Datapoint{12., 0, "test6", ""},
	}), nil))

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{3.3, 0, "test0", ""},
	}), nil))

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{3.3, 3, "test0", ""},
		&Datapoint{6.8, 0, "test0", ""},
	}), nil))

	cmpQuery(t, s, &Query{
		Timeseries: oid1,
	}, DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
		&Datapoint{3., 0, "test2", ""},
		&Datapoint{3.3, 3, "test0", ""},
		&Datapoint{6.8, 0, "test0", ""},
		&Datapoint{7., 0, "test6", ""},
		&Datapoint{8., 0, "test6", ""},
		&Datapoint{9., 0, "test6", ""},
		&Datapoint{10., 0, "test6", ""},
		&Datapoint{11., 0, "test6", ""},
		&Datapoint{12., 0, "test6", ""},
	})

	// Clear the timeseries
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
	}))
	l, err = s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	// Test conflicts in mergeBatch
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
		&Datapoint{3., 0, "test2", ""},
	}), nil))

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{3, 0, "different", ""},
	}), nil))

	cmpQuery(t, s, &Query{
		Timeseries: oid1,
	}, DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
		&Datapoint{3, 0, "different", ""},
	})

	// Clear the timeseries
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
	}))
	l, err = s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	// Test conflicts in mergeBatch - second type
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "test1", ""},
		&Datapoint{2.9, 1, "test2", ""},
	}), nil))

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{2, 0, "different", ""},
		&Datapoint{3, 0, "different", ""},
	}), nil))

	cmpQuery(t, s, &Query{
		Timeseries: oid1,
	}, DatapointArray{
		&Datapoint{1., 0, "test0", ""},
		&Datapoint{2., 0, "different", ""},
		&Datapoint{3, 0, "different", ""},
	})
}

func TestDelete(t *testing.T) {
	adb, oid1, _, cleanup := newDBWithObjects(t)
	defer cleanup()

	s := TimeseriesDB{
		DB:                    adb,
		BatchSize:             3,
		MaxBatchSize:          5,
		BatchCompressionLevel: 3,
	}

	dpa8 := DatapointArray{
		&Datapoint{1., .8, "test0", ""},
		&Datapoint{2., .7, "test1", ""},
		&Datapoint{3., 1, "test2", ""},
		&Datapoint{4., .5, "test3", ""},
		&Datapoint{5., .4, "test4", ""},
		&Datapoint{6., .2, "test6", ""},
		&Datapoint{7., .1, "test7", ""},
		&Datapoint{8., 0, "test8", ""},
		&Datapoint{9., 0, "test9", ""},
	}

	err := s.Insert(oid1, NewDatapointArrayIterator(dpa8), nil)
	require.NoError(t, err)
	cmpQuery(t, s, &Query{Timeseries: oid1}, dpa8)

	i1 := int64(1)
	i2 := int64(-1)
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
		T1:         float64(4),
		T2:         float64(7),
		I1:         &i1,
		I2:         &i2,
	}))

	l, err := s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(6), l)
	cmpQuery(t, s, &Query{Timeseries: oid1, T1: float64(4)}, dpa8[6:])
	cmpQuery(t, s, &Query{Timeseries: oid1, T2: float64(6)}, dpa8[:3])

	// Re-add the data
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(dpa8), nil))
	cmpQuery(t, s, &Query{Timeseries: oid1}, dpa8)

	// Try deleting again, this time a single datapoint at the boundary
	i1 = -7
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
		I:          &i1,
	}))
	l, err = s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(8), l)
	i1 = -6
	cmpQuery(t, s, &Query{Timeseries: oid1, I1: &i1}, dpa8[3:])
	cmpQuery(t, s, &Query{Timeseries: oid1, I2: &i1}, dpa8[:2])

	// Re-add the data
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(dpa8), nil))
	cmpQuery(t, s, &Query{Timeseries: oid1}, dpa8)

	// Delete by timestamp at boundary
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
		T:          float64(7),
	}))
	l, err = s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(8), l)
	i2 = 5000
	cmpQuery(t, s, &Query{Timeseries: oid1, T2: float64(7.5)}, dpa8[:6])
	cmpQuery(t, s, &Query{Timeseries: oid1, T1: float64(7)}, dpa8[7:])

	// Re-add the data
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(dpa8), nil))
	cmpQuery(t, s, &Query{Timeseries: oid1}, dpa8)

	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
	}))
	l, err = s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	// Re-add the data
	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(dpa8), nil))
	cmpQuery(t, s, &Query{Timeseries: oid1}, dpa8)

	i1 = -1000
	i2 = 1000
	require.NoError(t, s.Delete(&Query{
		Timeseries: oid1,
		I1:         &i1,
		I2:         &i2,
	}))
	l, err = s.Length(oid1, false)
	require.NoError(t, err)
	require.Equal(t, int64(0), l)
}

func TestBatching(t *testing.T) {
	adb, oid1, _, cleanup := newDBWithObjects(t)
	defer cleanup()

	s := TimeseriesDB{
		DB:                    adb,
		BatchSize:             3,
		MaxBatchSize:          5,
		BatchCompressionLevel: 3,
	}

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{1., .8, "test0", ""},
		&Datapoint{2., .7, "test1", ""},
		&Datapoint{3., 1, "test2", ""},
		&Datapoint{7., .1, "test7", ""},
		&Datapoint{8., 0, "test8", ""},
		&Datapoint{9., 0, "test9", ""},
		&Datapoint{10., .1, "test10", ""},
		&Datapoint{11., 0, "test11", ""},
		&Datapoint{12., 0, "test12", ""},
	}), nil))

	require.NoError(t, s.Insert(oid1, NewDatapointArrayIterator(DatapointArray{
		&Datapoint{4., .5, "test3", ""},
		&Datapoint{5., .4, "test4", ""},
		&Datapoint{6., 5.1, "test6", ""},
	}), nil))

	cmpQuery(t, s, &Query{
		Timeseries: oid1,
	}, DatapointArray{
		&Datapoint{1., .8, "test0", ""},
		&Datapoint{2., .7, "test1", ""},
		&Datapoint{3., 1, "test2", ""},
		&Datapoint{4., .5, "test3", ""},
		&Datapoint{5., .4, "test4", ""},
		&Datapoint{6., 5.1, "test6", ""},
		&Datapoint{12., 0, "test12", ""},
	})
}

func TestDurationUpdate(t *testing.T) {
	adb, oid1, _, cleanup := newDBWithObjects(t)
	defer cleanup()

	s := TimeseriesDB{
		DB:                    adb,
		BatchSize:             3,
		MaxBatchSize:          5,
		BatchCompressionLevel: 3,
	}

	insert1 := DatapointArray{
		&Datapoint{1., 1., 1, ""},
		&Datapoint{2., 1., 2, ""},
		&Datapoint{3., 1., 3, ""},
		&Datapoint{4., 1., 4, ""},
	}
	err := s.Insert(oid1, NewDatapointArrayIterator(insert1), nil)
	require.NoError(t, err)

	insert2 := DatapointArray{
		&Datapoint{2.5, 1., 2.5, ""},
		&Datapoint{3.5, 0, 3.5, ""},
	}
	err = s.Insert(oid1, NewDatapointArrayIterator(insert2), nil)
	require.NoError(t, err)

	di, err := s.Query(&Query{
		Timeseries: oid1,
	})
	require.NoError(t, err)
	dpa, err := NewArrayFromIterator(di)
	require.NoError(t, err)

	output := DatapointArray{
		&Datapoint{1., 1., 1, ""},
		&Datapoint{2.5, 1., 2.5, ""},
		&Datapoint{3.5, 0, 3.5, ""},
		&Datapoint{4., 1., 4, ""},
	}

	require.True(t, output.IsEqual(dpa), "%s different from %s", dpa.String(), output.String())
}
