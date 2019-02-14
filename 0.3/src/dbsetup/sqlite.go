/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"dbsetup/dbutil"

	log "github.com/Sirupsen/logrus"
)

//PostgresService is a service for running SQLite
type SqliteService struct {
	s *config.SQLService
}

func (s *SqliteService) Name() string {
	return "sqlite"
}

func (s *SqliteService) Status() Status {
	return StatusNone
}

func (s *SqliteService) Kill() error {
	// sqlite doesn't have a process :)
	return nil
}
func (s *SqliteService) Stop() error {
	// sqlite doesn't have a process :)
	return nil
}
func (s *SqliteService) Start() error {
	log.Infof("Using sqlite3 database")
	// sqlite doesn't have a process :)
	return nil
}

//Create prepares Postgres
func (s *SqliteService) Create() error {
	log.Infof("Setting up sqlite3 database")
	return dbutil.SetupDatabase("sqlite3", s.s.URI)
}

//NewPostgresService creates a new service for Postgres
func NewSqliteService(s *config.SQLService) *SqliteService {
	return &SqliteService{s}
}
