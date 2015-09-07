package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

/**

This file provides the main configuration system for ConnectorDB.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

const (
	// The database types we support
	Postgres = "postgres"

	// Localhost variables, both net and string
	LocalhostIpV4 = "127.0.0.1"
	LocalhostIpV6 = "0:0:0:0:0:0:0:1"

	ConfigFileName = "connectordb_config.json"

	currentConfigVersion = 0
)

var (
	configuration = newConfiguration()
	doneInit      bool

	ErrNotSetup = errors.New("InitConfiguration has not been called yet!")
)

type Configuration struct {
	ConfigVersion            int
	Nodetype                 string
	RedisHost                string
	RedisPort                int
	GnatsdHost               string
	GnatsdPort               int
	DatabaseConnectionString string
	WebPort                  int
	RunApi                   bool
	RunWeb                   bool
	RunDaisy                 bool
	PostgresPort             int
	PostgresHost             string
	DatabaseType             string
	StreamdbDirectory        string
	DisallowedNames          []string
}

// Gets the streamdb directory, failing if config hasn't been set up yet.
func GetStreamdbDirectory() (string, error) {
	if !doneInit {
		return "", ErrNotSetup
	}

	return configuration.StreamdbDirectory, nil
}

// Returns the database connection string for the current database
func GetDatabaseConnectionString() string {
	return configuration.GetDatabaseConnectionString()
}

// Returns the database connection string for the current database
func (config *Configuration) GetDatabaseConnectionString() string {
	return fmt.Sprintf("postgres://%v:%v/connectordb?sslmode=disable", configuration.PostgresHost, configuration.PostgresPort)
}

// Returns the redis "uri", no prefix appneded
func (config *Configuration) GetRedisUri() string {
	return fmt.Sprintf("%s:%d", configuration.RedisHost, configuration.RedisPort)
}

// Get the gnatsd "uri" no prefix appended; it'll be in the format host:port
func (config *Configuration) GetGnatsdUri() string {
	return fmt.Sprintf("%s:%d", configuration.GnatsdHost, configuration.GnatsdPort)
}

func newConfiguration() *Configuration {
	var cfg Configuration
	cfg.ConfigVersion = currentConfigVersion
	cfg.Nodetype = "master"
	cfg.RedisHost = LocalhostIpV4
	cfg.RedisPort = 6379
	cfg.GnatsdHost = LocalhostIpV4
	cfg.GnatsdPort = 4222
	cfg.WebPort = 8000
	cfg.RunApi = true
	cfg.RunWeb = true
	cfg.RunDaisy = false
	cfg.PostgresHost = LocalhostIpV4
	cfg.PostgresPort = 52592
	cfg.DatabaseType = Postgres
	cfg.DisallowedNames = []string{"postmaster", "root"}
	return &cfg
}

func loadConfigurationFromFile(cdbDirectory string) error {
	path := filepath.Join(cdbDirectory, ConfigFileName)

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	json.Unmarshal(file, configuration)
	return nil
}

func dumpConfigurationToFile(cdbDirectory string) error {
	path := filepath.Join(cdbDirectory, ConfigFileName)

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	json.Unmarshal(file, configuration)
	return nil
}

// Loads a configuration from a path if possible
func InitConfiguration(path string) error {
	if doneInit {
		return nil
	}
	err := loadConfigurationFromFile(path)

	configuration.StreamdbDirectory = path // save for saving
	doneInit = true

	return err
}

// Loads the configuration from the same place it was loaded from originally
func ReloadConfiguration() error {
	return loadConfigurationFromFile(configuration.StreamdbDirectory)
}

func GetConfiguration() *Configuration {
	return configuration
}

// Dumps the configuration to the place it was loaded from originally.
func SaveConfiguration() error {
	return dumpConfigurationToFile(configuration.StreamdbDirectory)
}

// TODO: add a saving daemon and check for reload signal/file changes
