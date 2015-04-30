package dbmaker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"streamdb/dbutil"
	"time"
	"streamdb/util"
)

var postgresDatabaseName = "postgres_database"


//GetSqlPath gets the connection string to the database
func GetSqlPath(streamdbDirectory, iface string, port int, err error) (string, error) {
	dbtype, err := GetDatabaseType(streamdbDirectory, err)
	if err != nil {
		return "", err
	}

	switch dbtype {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%d/connectordb?sslmode=disable", iface, port), nil
	case "sqlite":
		return filepath.Join(streamdbDirectory, sqliteDatabaseName), nil
	default:
		return "", ErrUnrecognizedDatabase
	}
}


//StartSqlDatabase starts the correct sql database based upon the directory
func StartSqlDatabase(streamdbDirectory, iface string, port int, err error) error {
	err = util.EnsurePidNotRunning(streamdbDirectory, "sqldb", err)

	dbtype, err := GetDatabaseType(streamdbDirectory, err)

	switch dbtype {
	case "postgres":
		err = StartPostgres(streamdbDirectory, iface, port, err)
	case "sqlite":
		err = StartSqlite(streamdbDirectory, iface, port, err)
	default:
		return ErrUnrecognizedDatabase
	}
	return err
}

//StopSqlDatabase stops the correct sql database based upon the directory
func StopSqlDatabase(streamdbDirectory string, err error) error {

	dbtype, err := GetDatabaseType(streamdbDirectory, err)

	switch dbtype {
	case "postgres":
		err = StopPostgres(streamdbDirectory, err)
	case "sqlite":
		err = StopSqlite(streamdbDirectory, err)
	default:
		return ErrUnrecognizedDatabase
	}
	return nil
}

//InitializePostgres creates a postgres database and subsequently sets it up to work with streamdb
func InitializePostgres(streamdbDirectory string, err error) error {
	if err != nil {
		return err
	}

	dbDir := filepath.Join(streamdbDirectory, postgresDatabaseName)
	log.Println("Setting up Postgres database")

	err = os.Mkdir(dbDir, FolderPermissions)

	//Now copy the configuration file
	err = CopyConfig(streamdbDirectory, "postgres.conf", err)

	//Initialize the database directory
	err = RunCommand(err, dbutil.FindPostgresInit(), "-D", dbDir)

	if err != nil {
		return err
	}

	//Now we create the underlying database
	err = StartPostgres(streamdbDirectory, "127.0.0.1", 55412, err)

	err = RunCommand(err, dbutil.FindPostgresPsql(), "-h", "localhost", "-p", "55412", "-d", "postgres", "-c", "CREATE DATABASE connectordb;")

	spath, err := GetSqlPath(streamdbDirectory, "127.0.0.1", 55412, err)
	if err == nil {
		log.Printf("Setting up initial tables\n")
		err = dbutil.UpgradeDatabase(spath, true)
	}

	StopPostgres(streamdbDirectory, nil)

	return err
}

//StartPostgres starts the postgres server for the database
func StartPostgres(streamdbDirectory, iface string, port int, err error) error {
	if err != nil {
		return err
	}
	log.Printf("Starting postgres server on port %d\n", port)
	configfile, err := SetConfig(streamdbDirectory, "postgres.conf",
		GenerateConfigReplacements(streamdbDirectory, "postgres", iface, port), err)

	postgresDir := filepath.Join(streamdbDirectory, postgresDatabaseName)

	//Postgres is a little bitch about its config file, which needs to be moved to the database dir
	err = util.CopyFileContents(configfile, filepath.Join(postgresDir, "postgresql.conf"), err)

	err = RunDaemon(err, dbutil.FindPostgres(), "-D", postgresDir)

	err = WaitPort(iface, port, err)
	if err == nil {
		//Sleep one second, since postgres is weird like that
		time.Sleep(1 * time.Second)
	}
	return err
}

//StopPostgres kills the postgres server associated with the database
func StopPostgres(streamdbDirectory string, err error) error {
	if err != nil {
		return err
	}
	log.Printf("Stopping postgres server\n")
	err = StopProcess(streamdbDirectory, "postgres", err)

	if err == nil {
		//Sleep a couple seconds, since postgres is weird like that
		time.Sleep(3 * time.Second)
	}
	return err
}
