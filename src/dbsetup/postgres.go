/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"dbsetup/dbutil"
	"os"
	"path/filepath"
	"strconv"
	"util"

	log "github.com/Sirupsen/logrus"
)

var (
	postgresDatabaseName = "postgres_database"
)

//PostgresService is a service for running Postgres
type PostgresService struct {
	BaseService
}

//Create prepares Postgres
func (s *PostgresService) Create() error {
	err := s.BaseService.Create()
	if err != nil {
		return err
	}
	dbDir := filepath.Join(s.ServiceDirectory, postgresDatabaseName)
	err = os.Mkdir(dbDir, FolderPermissions)
	if err != nil {
		return err
	}

	//Initialize the database directory
	err = util.RunCommand(err, GetPostgresExecutablePath("initdb"), "-D", dbDir)
	if err != nil {
		return err
	}

	//Now we need to start postgres so that we can create all the necessary tables
	if err = s.Start(); err != nil {
		return err
	}

	port := strconv.Itoa(int(s.S.Port))
	err = util.RunCommand(err, GetPostgresExecutablePath("psql"), "-h", s.S.Hostname, "-p", port, "-d", "postgres", "-c", "CREATE DATABASE connectordb;")
	if err != nil {
		return err
	}

	return dbutil.SetupDatabase("postgres", s.S.GetSqlConnectionString())
}

//Start starts the service
func (s *PostgresService) Start() error {
	postgresDir := filepath.Join(s.ServiceDirectory, postgresDatabaseName)
	postgresSettingsPath := filepath.Join(postgresDir, "postgresql.conf")

	configfile, err := s.start()

	//Postgres is a little bitch about its config file, which needs to be moved to the database dir
	err = util.CopyFileContents(configfile, postgresSettingsPath, err)
	if err != nil {
		return err
	}

	err = util.RunCommand(err, GetPostgresExecutablePath("pg_ctl"), "-D", postgresDir, "-w", "start")
	err = util.WaitPort(s.S.Hostname, int(s.S.Port), err)

	return err
}

//Stop shuts down the Postgres server
func (s *PostgresService) Stop() error {
	if s == nil {
		return nil
	}
	log.Print("Stopping Postgres...")
	postgresDir := filepath.Join(s.ServiceDirectory, postgresDatabaseName)

	return util.RunCommand(nil, GetPostgresExecutablePath("pg_ctl"), "-D", postgresDir, "-m", "fast", "stop")
}

//NewPostgresService creates a new service for Postgres
func NewPostgresService(serviceDirectory string, s *config.Service) *PostgresService {
	return &PostgresService{BaseService{serviceDirectory, "postgres", StatusNone, s}}
}
