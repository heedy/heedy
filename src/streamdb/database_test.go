package streamdb

import (
	"database/sql"
	"os"
	"testing"
)

func TableMakerTestCreate(db *sql.DB) error {
	db.Exec("DROP TABLE IF EXISTS timebatchtable")
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS timebatchtable
        (
            Key VARCHAR NOT NULL,
            EndTime BIGINT,
            EndIndex BIGINT,
			Version INTEGER,
            Data BYTEA,
            PRIMARY KEY (Key, EndIndex)
            );`)
	db.Exec("CREATE INDEX keytime ON timebatchtable (Key,EndTime ASC);")
	return err
}

func TestDatabaseOpen(t *testing.T) {
	os.Remove("TESTING_timebatch.db") //Delete sqlite database if exists
	sdb, err := sql.Open("sqlite3", "TESTING_timebatch.db")
	if err != nil {
		t.Errorf("Couldn't open database: %v", err)
		return
	}
	TableMakerTestCreate(sdb)
	sdb.Close()
	db, err := Open("sqlite://TESTING_timebatch.db", "localhost:6379", "localhost:4222")
	if err != nil {
		t.Errorf("Could not open streamdb: %s", err)
		return
	}
	defer db.Close()
}
