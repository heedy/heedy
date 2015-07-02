package datastream

import (
	"database/sql"

	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	log "github.com/Sirupsen/logrus"
)

var (
	ds  *DataStream
	rc  *RedisConnection
	sdb *SqlStore
	err error
)

func TestMain(m *testing.M) {
	sqldb, err := sql.Open("postgres", "sslmode=disable dbname=connectordb port=52592")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	_, err = sqldb.Exec(`CREATE TABLE IF NOT EXISTS datastream (
	    StreamId BIGINT NOT NULL,
		Substream VARCHAR,
	    EndTime DOUBLE PRECISION,
	    EndIndex BIGINT,
		Version INTEGER,
	    Data BYTEA,
	    UNIQUE (StreamId, Substream, EndIndex),
	    PRIMARY KEY (StreamId, Substream, EndIndex)
	    );`)

	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	ds, err = OpenDataStream(sqldb, &DefaultOptions)
	if err != nil {
		log.Error(err)
		os.Exit(3)
	}
	ds.Close()

	ds, err = OpenDataStream(sqldb, &DefaultOptions)
	if err != nil {
		log.Error(err)
		os.Exit(4)
	}

	rc = ds.redis
	sdb = ds.sqls

	res := m.Run()

	ds.Close()
	os.Exit(res)
}

func TestBasics(t *testing.T) {
	ds.Clear()

	require.NoError(t, ds.DeleteStream(1))

	i, err := ds.StreamLength(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	err = ds.Insert(1, "", dpa6, false)
	require.NoError(t, err)

	i, err = ds.StreamLength(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(5), i)

	require.NoError(t, ds.DeleteStream(1))

	i, err = ds.StreamLength(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

}

func TestInsert(t *testing.T) {
	ds.Clear()
	ds.batchsize = 4

	require.NoError(t, ds.Insert(1, "", dpa7, false))

	require.Error(t, ds.Insert(1, "", dpa4, false))
	require.NoError(t, ds.Insert(1, "", dpa4, true))

}
