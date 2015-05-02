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

// A service representing the postgres database
type PostgresService struct {
	ServiceHelper // We get stop, status, kill, and Name from this
	host string
	port int
	streamdbDirectory string
}

// Creates and returns a new postgres service in a pre-init state
// with default values loaded from config
func NewDefaultPostgresService() *PostgresService {
	return NewConfigPostgresService(config.GetConfiguration())
}

// Creates a postgres service from a given configuration
func NewConfigPostgresService(config *config.Configuration) *PostgresService {
	host := config.PostgresHost
	port := config.PostgresPort
	dir	 := config.StreamdbDirectory
	return NewPostgresService(host, port, dir)
}

// Creates and returns a new postgres service in a pre-init state
func NewPostgresService(host string, port int, streamdbDirectory string) *PostgresService {
	var ps PostgresService
	ps.host = host
	ps.port = port
	ps.streamdbDirectory = streamdbDirectory

	ps.InitServiceHelper(streamdbDirectory, "postgres")

	log.Printf("Creating new postgres service at %s:%d, %s using %p\n", host, port, streamdbDirectory, &ps)
	return &ps
}

// Setup postgres including the create database.
func (srv *PostgresService) Setup() error {

	dbDir := filepath.Join(srv.streamdbDirectory, postgresDatabaseName)
	log.Printf("Setting up postgres service at %s:%d, %s using %p\n", srv.host, srv.port, srv.streamdbDirectory, srv)

	err := os.Mkdir(dbDir, FolderPermissions)

	//Now copy the configuration file
	err = CopyConfig(srv.streamdbDirectory, "postgres.conf", err)

	//Initialize the database directory
	err = RunCommand(err, dbutil.FindPostgresInit(), "-D", dbDir)

	if err != nil {
		return err
	}

	if err := srv.Init(); err != nil {
		return err
	}

	//Now we create the underlying database
	if err := srv.Start(); err != nil {
		return err
	}

	port := strconv.Itoa(srv.port)
	err = RunCommand(err, dbutil.FindPostgresPsql(), "-h", srv.host, "-p", port, "-d", "postgres", "-c", "CREATE DATABASE connectordb;")
	if err != nil {
		return err
	}

	log.Printf("Setting up initial tables\n")
	spath := config.GetDatabaseConnectionString()
	return dbutil.UpgradeDatabase(spath, true)
}

// Init the postgres service
func (srv *PostgresService) Init() error {
	log.Printf("Initializing Postgres\n")
	srv.Stat = StatusInit

	// Nothing to do here, may want to which/look for the executables in the
	// future and/or make sure the whole database is there

	return nil
}

// Starts postgres
func (srv *PostgresService) Start() error {
	if srv.Stat == StatusRunning {
		return nil
	}
	if srv.Stat != StatusInit {
		log.Printf("Could not start postgres, status is %v\n", srv.Stat)
		return ErrNotInitialized
	}
	srv.Stat = StatusRunning


	log.Printf("Starting postgres server on port %d\n", srv.port)
	postgresDir := filepath.Join(srv.streamdbDirectory, postgresDatabaseName)
	postgresSettingsPath := filepath.Join(postgresDir, "postgresql.conf")

	log.Printf("Postgres Directory: %s\n", postgresDir)
	log.Printf("Postgres Settings Path: %s\n", postgresSettingsPath)


	configReplacements := GenerateConfigReplacements(srv.streamdbDirectory, "postgres", srv.host, srv.port)

	configfile, err := SetConfig(srv.streamdbDirectory, "postgres.conf", configReplacements, nil)


	//Postgres is a little bitch about its config file, which needs to be moved to the database dir
	err = util.CopyFileContents(configfile, postgresSettingsPath, err)
	if err != nil {
		return err
	}


	err = RunDaemon(err, dbutil.FindPostgres(), "-D", postgresDir)

	err = WaitPort(srv.host, srv.port, err)
	if err != nil {
		return err
	}

	//Sleep one second, since postgres is weird like that
	time.Sleep(1 * time.Second)
	return nil
}

func (srv *PostgresService) Stop() error {
	err := srv.HelperStop()
	if err == nil {
		//Sleep a couple seconds, since postgres is weird like that
		time.Sleep(5 * time.Second)
	}
	return err
}


func (srv *PostgresService) Kill() error {
	return srv.HelperKill()
}
