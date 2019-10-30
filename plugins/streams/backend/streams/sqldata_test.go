package streams

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func genDatabase(t *testing.T) (*sqlx.DB, func()) {
	os.RemoveAll("./test_db")
	os.Mkdir("./test_db", 0755)
	db, err := sqlx.Open("sqlite3", "test_db/heedy.db?_fk=1")
	require.NoError(t, err)

	_, err = db.Exec(`
	CREATE TABLE sources (
		id VARCHAR(36) PRIMARY KEY
	);

	INSERT INTO sources VALUES ('s1'), ('s2');
`)
	require.NoError(t, err)

	return db, func() {
		os.RemoveAll("./test_db")
	}
}

func TestDatabase(t *testing.T) {
	sdb, cleanup := genDatabase(t)
	defer cleanup()
	require.NoError(t, CreateSQLData(sdb))
	action := true
	s := OpenSQLData(sdb)

	l, err := s.StreamDataLength("s1", false)
	require.NoError(t, err)
	require.Equal(t, l, uint64(0))

	require.NoError(t, s.WriteStreamData("s1", NewDatapointArrayIterator(dpa1), &InsertQuery{
		Actions: &action,
	}))

	tt := float64(1.0)
	di, err := s.ReadStreamData("s1", &Query{
		T:       &tt,
		Actions: &action,
	})
	require.NoError(t, err)
	dpa, err := NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 1)
	require.Equal(t, dpa.String(), dpa1[0:1].String())

	di, err = s.ReadStreamData("s1", &Query{
		T1:      &tt,
		Actions: &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 2)
	require.Equal(t, dpa.String(), dpa1.String())

	// Overwrite the first datapoint
	insertType := "upsert"
	require.NoError(t, s.WriteStreamData("s1", NewDatapointArrayIterator(dpa3), &InsertQuery{
		Type:    &insertType,
		Actions: &action,
	}))

	di, err = s.ReadStreamData("s1", &Query{
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

	l, err = s.StreamDataLength("s1", true)
	require.NoError(t, err)
	require.Equal(t, uint64(2), l)

	require.NoError(t, s.RemoveStreamData("s1", &Query{
		T:       &tt,
		Actions: &action,
	}))

	di, err = s.ReadStreamData("s1", &Query{
		T1:      &tt,
		Actions: &action,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa), 1)
	require.Equal(t, dpa[0].String(), dpa1[1].String())

	require.Error(t, s.WriteStreamData("s2", NewDatapointArrayIterator(dpa7), &InsertQuery{}))
	itype := "upsert"
	require.NoError(t, s.WriteStreamData("s2", NewDatapointArrayIterator(dpa7), &InsertQuery{
		Type: &itype,
	}))

	l, err = s.StreamDataLength("s2", false)
	require.NoError(t, err)
	require.Equal(t, uint64(len(dpa7)-1), l) // dpa7 has timestamp 6 repeated

	i1 := int64(1)
	i2 := int64(-3)
	di, err = s.ReadStreamData("s2", &Query{
		I1: &i1,
		I2: &i2,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, dpa.String(), dpa7[1:5].String())

	i2 = 80
	di, err = s.ReadStreamData("s2", &Query{
		I2: &i2,
	})
	require.NoError(t, err)
	dpa, err = NewArrayFromIterator(di)
	require.NoError(t, err)
	require.Equal(t, len(dpa7)-1, len(dpa))

}
