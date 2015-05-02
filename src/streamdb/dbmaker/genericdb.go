package dbmaker
/**

import (
	"streamdb/dbutil"
	"streamdb/util"
	"streamdb/config"
	"errors"
)
var (
	ErrNotLocalDb = errors.New("the specified database is not on this host")
)

//StartSqlDatabase starts the correct sql database based upon the directory
func StartSqlDatabase() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	err = util.EnsurePidNotRunning(streamdbDirectory, "sqldb", nil)
	if err != nil {
		return err
	}

	cs := config.GetDatabaseConnectionString()
	_, dbtype := dbutil.ProcessConnectionString(cs)

	if ! config.IsDatabaseLocal() {
		return ErrNotLocalDb
	}

	switch dbtype {
	case "postgres":
		err = StartPostgres()
	case "sqlite":
		err = StartSqlite()
	default:
		return ErrUnrecognizedDatabase
	}
	return err
}

//StopSqlDatabase stops the correct sql database based upon the directory
func StopSqlDatabase() error {
	cs := config.GetConfiguration().DatabaseConnectionString
	_, dbtype := dbutil.ProcessConnectionString(cs)

	switch dbtype {
	case "postgres":
		return StopPostgres()
	case "sqlite":
		return StopSqlite()
	default:
		return ErrUnrecognizedDatabase
	}
	return nil
}
**/
