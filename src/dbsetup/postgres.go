package dbsetup

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
	"util"
	"util/dbutil"

	log "github.com/Sirupsen/logrus"
)

var (
	postgresDatabaseName = "postgres_database"
)

//PostgresService is a service for running Postgres
type PostgresService struct {
	BaseService
	Host             string
	Port             int
	ConnectionString string
}

//Create prepares Postgres
func (s *PostgresService) Create() error {
	log.Infof("Setting up Postgres server")
	dbDir := filepath.Join(s.ServiceDirectory, postgresDatabaseName)

	err := os.Mkdir(dbDir, FolderPermissions)
	err = CopyConfig(s.ServiceDirectory, "postgres.conf", err)

	//Initialize the database directory
	err = util.RunCommand(err, dbutil.FindPostgresInit(), "-D", dbDir)
	if err != nil {
		return err
	}

	//Now we need to start postgres so that we can create all the necessary tables
	if err = s.Start(); err != nil {
		return err
	}

	port := strconv.Itoa(s.Port)
	err = util.RunCommand(err, dbutil.FindPostgresPsql(), "-h", s.Host, "-p", port, "-d", "postgres", "-c", "CREATE DATABASE connectordb;")
	if err != nil {
		return err
	}

	return dbutil.UpgradeDatabase(s.ConnectionString, true)
}

//Start starts the service
func (s *PostgresService) Start() error {
	if s.Status() == StatusRunning {
		return nil
	}
	s.Stat = StatusError

	log.Infof("Staring postgres on port %d", s.Port)
	postgresDir := filepath.Join(s.ServiceDirectory, postgresDatabaseName)
	postgresSettingsPath := filepath.Join(postgresDir, "postgresql.conf")

	configReplacements := GenerateConfigReplacements(s.ServiceDirectory, "postgres", s.Host, s.Port)
	configfile, err := SetConfig(s.ServiceDirectory, "postgres.conf", configReplacements, nil)

	//Postgres is a little bitch about its config file, which needs to be moved to the database dir
	err = util.CopyFileContents(configfile, postgresSettingsPath, err)
	if err != nil {
		return err
	}

	err = util.RunDaemon(err, dbutil.FindPostgres(), "-D", postgresDir)
	err = util.WaitPort(s.Host, s.Port, err)

	if err == nil {
		s.Stat = StatusRunning

		//Sleep one second, since postgres is weird like that
		time.Sleep(1 * time.Second)
	}

	return err
}

//Stop shuts down the Postgres server
func (s *PostgresService) Stop() error {
	log.Print("Stopping Postgres...")
	pgctl := dbutil.FindPostgresPgctl()
	postgresDir := filepath.Join(s.ServiceDirectory, postgresDatabaseName)

	return util.RunCommand(nil, pgctl, "-D", postgresDir, "-m", "fast", "stop")
}

//NewPostgresService creates a new service for Postgres
func NewPostgresService(serviceDirectory, connectionstring, host string, port int) *PostgresService {
	return &PostgresService{BaseService{serviceDirectory, "postgres", StatusNone}, host, port, connectionstring}
}
