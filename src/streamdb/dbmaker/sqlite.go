package dbmaker

import (
	"log"
	"path/filepath"
	"streamdb/dbutil"
)

var sqliteDatabaseName = "streamdb.sqlite3"

//InitializeSqlite creates an sqlite database and subsequently sets it up to work with streamdb
func InitializeSqlite(streamdbDirectory string, err error) error {
	if err != nil {
		return err
	}

	dbFile := filepath.Join(streamdbDirectory, sqliteDatabaseName)
	log.Printf("Initializing sqlite database '%s'\n", dbFile)

	// because sqlite doesn't always like being started on a file that
	// doesn't exist
	Touch(dbFile)
	//Initialize the database tables
	log.Printf("Setting up initial tables\n")
	return dbutil.UpgradeDatabase(dbFile, true)
}

//StartSqlite does absolutely nothing, since sqlite is a single-process thing
func StartSqlite(streamdbDirectory, iface string, port int, err error) error {
	return err
}

//StopSqlite does absolutely nothing, since sqlite is a single-process thing
func StopSqlite(streamdbDirectory string, err error) error {
	return err
}
