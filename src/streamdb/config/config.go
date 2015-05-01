package config

import (
	"net"
	"errors"
	"bytes"
	"fmt"
)

/**

This file provides the main configuration system for ConnectorDB.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
)

var (
	configuration = newConfiguration()
	doneInit bool

	ErrNotSetup = errors.New("InitConfiguration has not been called yet!")
)

type Configuration struct {
	Nodetype string
	RedisHost string
	RedisPort int
	GnatsdHost string
	GnatsdPort int
	DatabaseConnectionString string
	WebPort int
	RunApi bool
	RunWeb bool
	RunDaisy bool
	PostgresPort int
	PostgresHost string
	SqliteDbPath string
	DatabaseType string
	StreamdbDirectory string

}

// Gets the streamdb directory, failing if config hasn't been set up yet.
func GetStreamdbDirectory() (string, error) {
	if ! doneInit {
		return "", ErrNotSetup
	}

	return configuration.StreamdbDirectory, nil
}

// Returns the database connection string for the current database
func GetDatabaseConnectionString() string {
	if configuration.DatabaseType == "sqlite" {
		return "sqlite://" + configuration.StreamdbDirectory  + "/" + configuration.SqliteDbPath
	}

	return fmt.Sprintf("postgres://%v:%v/connectordb?sslmode=disable", configuration.PostgresHost, configuration.PostgresPort)
}

// Checks if a database needs to be started lcoally
func IsDatabaseLocal() bool {
	if configuration.DatabaseType == "sqlite" {
		return true
	}

	ips, err := net.LookupIP(configuration.PostgresHost)

	if err != nil {
		return true // another db will catch if there's a problem
	}

	localV4 := net.ParseIP("127.0.0.1")
	localV6 := net.ParseIP("0:0:0:0:0:0:0:1")

	for _, ip := range(ips) {
		if bytes.Compare(ip, localV4) == 0 || bytes.Compare(ip, localV6) == 0 {
			return true
		}
	}

	return false
}

func newConfiguration() *Configuration {
	var cfg Configuration
	cfg.Nodetype = "master"
	cfg.RedisHost = "localhost"
	cfg.RedisPort = 6379
	cfg.GnatsdHost = "localhost"
	cfg.GnatsdPort = 4222
	cfg.WebPort = 8000
	cfg.RunApi = true
	cfg.RunWeb = true
	cfg.RunDaisy = false
	cfg.SqliteDbPath = ""
	cfg.PostgresHost = "localhost"
	cfg.PostgresPort = 52592
	cfg.DatabaseType = "postgres"

	return &cfg
}


// Loads a configuration from a path if possible
func InitConfiguration(path string) error {
	if doneInit {
		return nil
	}
	doneInit = true

	configuration.StreamdbDirectory = path // save for saving

	// TODO load config from here if possible
	return nil
}

func ReloadConfiguration() {

}

func GetConfiguration() (*Configuration) {
	return configuration
}


func SaveConfiguration() error {
	return nil
}

//TODO: add a saving daemon and check for reload signal/file changes
