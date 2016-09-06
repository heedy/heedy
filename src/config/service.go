/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

type Service struct {
	Hostname string `json:"hostname"`
	Port     uint16 `json:"port"`

	//Username and password are used for login to constituent servers
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	Enabled bool `json:"enabled"` //Whether or not to run the service on "connectordb start"
}

// GetRedisConnectionString returns the string used to connect to redis
func (s *Service) GetRedisConnectionString() string {
	return fmt.Sprintf("%s:%d", s.Hostname, s.Port)
}

// GetNatsConnectionString returns the string used to connect to NATS
func (s *Service) GetNatsConnectionString() string {
	return fmt.Sprintf("nats://%s:%s@%s:%d", s.Username, s.Password, s.Hostname, s.Port)
}

// GetSqlConnectionString checks server type and returns either the filename or postgres url
func (s *Service) GetSqlConnectionString() string {
	if s.Password == "" {
		return fmt.Sprintf("postgres://%v:%v/connectordb?sslmode=disable", s.Hostname, s.Port)
	}
	return fmt.Sprintf("postgres://%v:%v@%v:%v/connectordb?sslmode=disable", s.Username, s.Password, s.Hostname, s.Port)

}

type SQLService struct {
	URI  string `json:"uri"`
	Type string `json:"type"` // The sql database type

	Service
}

// GetSqlConnectionString checks server type and returns either the filename or postgres url
func (s *SQLService) GetSqlConnectionString() string {
	// If there is a uri given, use that
	if s.URI != "" {
		return s.URI
	}
	return s.Service.GetSqlConnectionString()
}

func (s *SQLService) Validate() (err error) {
	if s.Type == "" {
		s.Type = "postgres"
	}
	if s.Type != "postgres" && s.Type != "sqlite3" {
		return errors.New("Unrecognized sql database type")
	}
	if s.Type == "sqlite3" {
		// If the database is sqlite, we have to set the filename up
		if s.URI == "" {
			s.URI = "db.sqlite3"
		}

		s.URI, err = filepath.Abs(s.URI)
		if err != nil {
			return err
		}
	} else {
		// The default postgres user is postgres
		if s.Username == "" {
			s.Username = "postgres"
		}
	}
	return nil

}
