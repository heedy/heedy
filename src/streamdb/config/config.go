package config

/**

This file provides the main configuration system for ConnectorDB.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"path/filepath"

	"github.com/kardianos/osext"
)

var (
	config *Configuration
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
}

func NewConfiguration() Configuration {
	var cfg Configuration
	cfg.Nodetype = "master"
	cfg.RedisHost = "localhost"
	cfg.RedisPort = 6379
	cfg.GnatsdHost = "localhost"
	cfg.GnatsdPort = 4222
	cfg.DatabaseConnectionString = ""
	cfg.WebPort = 8000
	cfg.RunApi = true
	cfg.RunWeb = true
	cfg.RunDaisy = false
	cfg.PostgresPort = 52592

	return cfg
}

func InitConfiguration(path string) error {
	return nil
}

func ReloadConfiguration() {

}

func GetConfiguration() (*Configuration, error) {
	return config, nil
}


func SaveConfiguration() error {
	return nil
}


//ConfigPath returns the path to the default StreamDB config templates
func ConfigPath() (string, error) {

	execpath, err := osext.ExecutableFolder()
	return filepath.Join(execpath, "config"), err
}


// Gets whether this is a postgres or sqlite connection
func (c Configuration) GetConnectionType() {
	
}
