package dbmaker

import (
	"log"
	"path/filepath"
	"streamdb/dbutil"
	"streamdb/util"
	"streamdb/config"
)

var sqliteDatabaseName = "streamdb.sqlite3"

//InitializeSqlite creates an sqlite database and subsequently sets it up to work with streamdb
func InitializeSqlite() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	dbFile := filepath.Join(streamdbDirectory, sqliteDatabaseName)
	log.Printf("Initializing sqlite database '%s'\n", dbFile)

	// because sqlite doesn't always like being started on a file that
	// doesn't exist
	util.Touch(dbFile)

	//Initialize the database tables
	log.Printf("Setting up initial tables\n")
	return dbutil.UpgradeDatabase(dbFile, true)
}

//StartSqlite does absolutely nothing, since sqlite is a single-process thing
func StartSqlite() error {
	return nil
}

//StopSqlite does absolutely nothing, since sqlite is a single-process thing
func StopSqlite() error {
	return nil
}
