package timeseries

import (
	"os"
	"testing"

	"github.com/heedy/heedy/backend/database"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func genDatabase(t *testing.T) (*database.AdminDB, func()) {
	os.RemoveAll("./test_db")
	os.Mkdir("./test_db", 0755)
	db, err := sqlx.Open("sqlite3", "test_db/heedy.db?_fk=1")
	require.NoError(t, err)

	_, err = db.Exec(`
	CREATE TABLE objects (
		id VARCHAR(36) PRIMARY KEY
	);

	INSERT INTO objects VALUES ('s1'), ('s2');
`)
	require.NoError(t, err)

	adb := &database.AdminDB{}
	adb.SqlxCache.InitCache(db)

	return adb, func() {
		os.RemoveAll("./test_db")
	}
}

func TestDatabase(t *testing.T) {
	sdb, cleanup := genDatabase(t)
	defer cleanup()
	require.NoError(t, SQLUpdater(sdb, nil, 0))
	action := true
	s := OpenSQLData(sdb)

	l, err := s.TimeseriesDataLength("s1", false)
	require.NoError(t, err)
	require.Equal(t, l, uint64(0))

	imethod := "insert"
	_, _, _, _, err = s.WriteTimeseriesData("s1", NewDatapointArrayIterator(dpa1), &InsertQuery{
		Actions: &action,
		Method:  &imethod,
	})
	require.NoError(t, err)

	tt := "1.0"
	di, err := s.ReadTimeseriesData("s1", &Query{
		T:       &tt,
		Actions: &action,
	})
	require.NoError(t, err)
	dpa, err := NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 1)
	require.Equal(t, dpa.String(), dpa1[0:1].String())

	di, err = s.ReadTimeseriesData("s1", &Query{
		T1:      &tt,
		Actions: &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 2)
	require.Equal(t, dpa.String(), dpa1.String())

	// Overwrite the first datapoint
	insertType := "update"
	_, _, _, _, err = s.WriteTimeseriesData("s1", NewDatapointArrayIterator(dpa3), &InsertQuery{
		Method:  &insertType,
		Actions: &action,
	})
	require.NoError(t, err)

	di, err = s.ReadTimeseriesData("s1", &Query{
		T1:      &tt,
		Actions: &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 2)
	require.NotEqual(t, dpa.String(), dpa1.String())
	require.Equal(t, dpa[0].String(), dpa3[0].String())
	require.Equal(t, dpa[1].String(), dpa1[1].String())

	l, err = s.TimeseriesDataLength("s1", true)
	require.NoError(t, err)
	require.Equal(t, uint64(2), l)

	require.NoError(t, s.RemoveTimeseriesData("s1", &Query{
		T:       &tt,
		Actions: &action,
	}))

	di, err = s.ReadTimeseriesData("s1", &Query{
		T1:      &tt,
		Actions: &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 1)
	require.Equal(t, dpa[0].String(), dpa1[1].String())

	_, _, _, _, err = s.WriteTimeseriesData("s2", NewDatapointArrayIterator(dpa7), &InsertQuery{
		Method: &imethod,
	})
	require.Error(t, err)
	itype := "update"
	_, _, _, _, err = s.WriteTimeseriesData("s2", NewDatapointArrayIterator(dpa7), &InsertQuery{
		Method: &itype,
	})
	require.NoError(t, err)

	l, err = s.TimeseriesDataLength("s2", false)
	require.NoError(t, err)
	require.Equal(t, uint64(len(dpa7)-1), l) // dpa7 has timestamp 6 repeated

	i1 := int64(1)
	i2 := int64(-3)
	di, err = s.ReadTimeseriesData("s2", &Query{
		I1: &i1,
		I2: &i2,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, dpa.String(), dpa7[1:5].String())

	i2 = 80
	di, err = s.ReadTimeseriesData("s2", &Query{
		I2: &i2,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa7)-1, len(dpa))

}

func TestDurationUpdate(t *testing.T) {
	sdb, cleanup := genDatabase(t)
	defer cleanup()
	require.NoError(t, SQLUpdater(sdb, nil, 0))

	s := OpenSQLData(sdb)

	insert1 := DatapointArray{
		&Datapoint{1., 1., 1, ""},
		&Datapoint{2., 1., 2, ""},
		&Datapoint{3., 1., 3, ""},
		&Datapoint{4., 1., 4, ""},
	}
	_, _, _, _, err := s.WriteTimeseriesData("s1", NewDatapointArrayIterator(insert1), &InsertQuery{})
	require.NoError(t, err)

	insert2 := DatapointArray{
		&Datapoint{2.5, 1., 2.5, ""},
		&Datapoint{3.5, 0, 3.5, ""},
	}
	_, _, _, _, err = s.WriteTimeseriesData("s1", NewDatapointArrayIterator(insert2), &InsertQuery{})
	require.NoError(t, err)

	di, err := s.ReadTimeseriesData("s1", &Query{})
	require.NoError(t, err)
	dpa, err := NewArrayFromIterator(di)
	require.NoError(t, err)

	output := DatapointArray{
		&Datapoint{1., 1., 1, ""},
		&Datapoint{2., .5, 2, ""},
		&Datapoint{2.5, 1., 2.5, ""},
		&Datapoint{3.5, 0, 3.5, ""},
		&Datapoint{4., 1., 4, ""},
	}

	require.Equal(t, output.String(), dpa.String())
}
