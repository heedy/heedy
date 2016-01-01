/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import "fmt"

type Service struct {
	Hostname string `json:"hostname"`
	Port     uint16 `json:"port"`

	//Username and password are used for login to constituent servers
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	//SSLPort uint16 `json:"sslport"` //The port on which to run Stunnel

	Enabled bool `json:"enabled"` //Whether or not to run the service on "connectordb start"
}

// GetSqlConnectionString assumes that the service is a postgres server, and returns the string url that can
// be used to connect to the server.
func (s *Service) GetSqlConnectionString() string {
	return fmt.Sprintf("postgres://%v:%v/connectordb?sslmode=disable", s.Hostname, s.Port)
}

// GetRedisConnectionString returns the string used to connect to redis
func (s *Service) GetRedisConnectionString() string {
	return fmt.Sprintf("%s:%d", s.Hostname, s.Port)
}

// GetNatsConnectionString returns the string used to connect to NATS
func (s *Service) GetNatsConnectionString() string {
	return fmt.Sprintf("nats://%s:%s@%s:%d", s.Username, s.Password, s.Hostname, s.Port)
}
