package services

import (
	"connectordb/config"
	"connectordb/streamdb/dbutil"
	"connectordb/streamdb/util"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	postgresDatabaseName = "postgres_database"
)

// A service representing the postgres database
type PostgresService struct {
	ServiceHelper     // We get stop, status, kill, and Name from this
	host              string
	port              int
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
	dir := config.StreamdbDirectory
	return NewPostgresService(host, port, dir)
}

// Creates and returns a new postgres service in a pre-init state
func NewPostgresService(host string, port int, streamdbDirectory string) *PostgresService {
	var ps PostgresService
	ps.host = host
	ps.port = port
	ps.streamdbDirectory = streamdbDirectory

	ps.InitServiceHelper(streamdbDirectory, "postgres")

	log.Debugf("Creating new postgres service at %s:%d, %s using %p", host, port, streamdbDirectory, &ps)
	return &ps
}

// Setup postgres including the create database.
func (srv *PostgresService) Setup() error {

	dbDir := filepath.Join(srv.streamdbDirectory, postgresDatabaseName)
	log.Printf("Setting up postgres service at %s:%d, %s using %p", srv.host, srv.port, srv.streamdbDirectory, srv)

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

	log.Debugf("Setting up initial tables")
	spath := config.GetDatabaseConnectionString()
	return dbutil.UpgradeDatabase(spath, true)
}

// Init the postgres service
func (srv *PostgresService) Init() error {
	log.Debugf("Initializing Postgres")
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
		log.Errorf("Could not start postgres, status is %v", srv.Stat)
		return ErrNotInitialized
	}
	srv.Stat = StatusRunning

	log.Printf("Starting postgres server on port %d", srv.port)
	postgresDir := filepath.Join(srv.streamdbDirectory, postgresDatabaseName)
	postgresSettingsPath := filepath.Join(postgresDir, "postgresql.conf")

	log.Debugf("Postgres Directory: %s", postgresDir)
	log.Debugf("Postgres Settings Path: %s", postgresSettingsPath)

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
	log.Print("Stopping postgres...")
	pgctl := dbutil.FindPostgresPgctl()
	postgresDir := filepath.Join(srv.streamdbDirectory, postgresDatabaseName)

	return RunCommand(nil, pgctl, "-D", postgresDir, "stop") //"-m", "fast", "stop")
}

func (srv *PostgresService) Kill() error {
	return srv.HelperKill()
}
