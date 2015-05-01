package dbmaker

import (
	"log"
	"os"
	"path/filepath"
	"streamdb/dbutil"
	"time"
	"streamdb/util"
	"streamdb/config"
	"strconv"
)

var (
	postgresDatabaseName = "postgres_database"
)

//InitializePostgres creates a postgres database and subsequently sets it up to work with streamdb
func InitializePostgres() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
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
	if err := StartPostgres(); err != nil {
		return err
	}
	defer StopPostgres()

	postgresHost := config.GetConfiguration().PostgresHost
	postgresPort := config.GetConfiguration().PostgresPort

	port := strconv.Itoa(postgresPort)


	err = RunCommand(err, dbutil.FindPostgresPsql(), "-h", postgresHost, "-p", port, "-d", "postgres", "-c", "CREATE DATABASE connectordb;")
	if err != nil {
		return err
	}

	log.Printf("Setting up initial tables\n")
	spath := config.GetDatabaseConnectionString()
	return dbutil.UpgradeDatabase(spath, true)
}

//StartPostgres starts the postgres server for the database
func StartPostgres() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	postgresHost := config.GetConfiguration().PostgresHost
	postgresPort := config.GetConfiguration().PostgresPort

	log.Printf("Starting postgres server on port %d\n", postgresPort)
	configfile, err := SetConfig(streamdbDirectory, "postgres.conf",
		GenerateConfigReplacements(streamdbDirectory, "postgres", postgresHost, postgresPort), err)

	postgresDir := filepath.Join(streamdbDirectory, postgresDatabaseName)

	//Postgres is a little bitch about its config file, which needs to be moved to the database dir
	err = util.CopyFileContents(configfile, filepath.Join(postgresDir, "postgresql.conf"), err)

	err = RunDaemon(err, dbutil.FindPostgres(), "-D", postgresDir)

	err = WaitPort(postgresHost, postgresPort, err)
	if err == nil {
		//Sleep one second, since postgres is weird like that
		time.Sleep(1 * time.Second)
	}
	return err
}

//StopPostgres kills the postgres server associated with the database
func StopPostgres() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
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
