package datastream

import (
	"database/sql"

	"os"
	"testing"

	_ "github.com/lib/pq"

	log "github.com/Sirupsen/logrus"
)

var (
	rc  *RedisConnection
	sdb *SqlStore
	err error
)

func TestMain(m *testing.M) {
	rc, err = NewRedisConnection(&DefaultOptions)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	rc.Close()

	rc, err = NewRedisConnection(&DefaultOptions)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	sqldb, err := sql.Open("postgres", "sslmode=disable dbname=connectordb port=52592")
	if err != nil {
		log.Error(err)
		rc.Close()
		os.Exit(2)
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
		rc.Close()
		sqldb.Close()
		os.Exit(3)
	}

	sdb, err = OpenSqlStore(sqldb)
	if err != nil {
		log.Error(err)
		rc.Close()
		sqldb.Close()
		os.Exit(4)
	}
	sdb.Close()

	sdb, err = OpenSqlStore(sqldb)
	if err != nil {
		log.Error(err)
		rc.Close()
		sqldb.Close()
		os.Exit(4)
	}

	res := m.Run()

	rc.Close()
	sdb.Close()
	sqldb.Close()
	os.Exit(res)
}
