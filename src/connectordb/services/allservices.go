package services

import (
	"connectordb/config"
	"connectordb/streamdb/util"
	"errors"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

var (
	sqliteInstance   *SqliteService
	postgresInstance *PostgresService
	gnatsdInstance   *GnatsdService
	redisInstance    *RedisService

	ErrNotInitialized     = errors.New("Module not yet initialized")
	ErrAlreadyInitialized = errors.New("All subsystems have allready been initialized")
	doneInit              = false
)

func initSqlDatabase(configuration *config.Configuration) error {
	if doneInit {
		return ErrAlreadyInitialized
	}

	sqliteInstance = NewConfigSqliteSerivce(configuration)
	postgresInstance = NewConfigPostgresService(configuration)

	sqlDatabaseType := configuration.DatabaseType

	switch sqlDatabaseType {
	case config.Postgres:
		if err := postgresInstance.Init(); err != nil {
			return err
		}
	case config.Sqlite:
		if err := sqliteInstance.Init(); err != nil {
			return err
		}
	default:
		return ErrUnrecognizedDatabase
	}

	return nil
}

func Init(configuration *config.Configuration) error {
	if doneInit {
		return ErrAlreadyInitialized
	}

	log.Debugf("Initializing subsystems")
	gnatsdInstance = NewConfigGnatsdService(configuration)
	redisInstance = NewConfigRedisService(configuration)

	if err := initSqlDatabase(configuration); err != nil {
		return err
	}

	if err := gnatsdInstance.Init(); err != nil {
		return err
	}

	if err := redisInstance.Init(); err != nil {
		return err
	}
	log.Debugf("Finished initializing subsystems")

	doneInit = true
	return nil
}

func startSqlDatabase(configuration *config.Configuration) error {
	sqlDatabaseType := configuration.DatabaseType

	switch sqlDatabaseType {
	case config.Postgres:
		if err := postgresInstance.Start(); err != nil {
			return err
		}
	case config.Sqlite:
		if err := sqliteInstance.Start(); err != nil {
			return err
		}
	default:
		return ErrUnrecognizedDatabase
	}

	return nil
}

//Start the necessary servers to run StreamDB
func Start(configuration *config.Configuration) error {
	log.Debugf("Starting subsystems")

	if err := startSqlDatabase(configuration); err != nil {
		return err
	}

	if err := gnatsdInstance.Start(); err != nil {
		return err
	}

	if err := redisInstance.Start(); err != nil {
		return err
	}

	util.Touch(filepath.Join(configuration.StreamdbDirectory, "connectordb.pid"))

	return nil
}

func stopSqlDatabase(configuration *config.Configuration) error {
	sqlDatabaseType := configuration.DatabaseType

	switch sqlDatabaseType {
	case config.Postgres:
		return postgresInstance.Stop()
	case config.Sqlite:
		return sqliteInstance.Stop()
	}
	return ErrUnrecognizedDatabase
}

//Start the necessary servers to run StreamDB
func Stop(configuration *config.Configuration) error {
	log.Debugf("Stopping subsystems")

	var globerr error
	if err := stopSqlDatabase(configuration); err != nil {
		globerr = err
	}

	if err := gnatsdInstance.Stop(); err != nil {
		globerr = err
	}

	if err := redisInstance.Stop(); err != nil {
		globerr = err
	}

	pidpath := filepath.Join(configuration.StreamdbDirectory, "connectordb.pid")
	if util.PathExists(pidpath) {
		if err := os.Remove(pidpath); err != nil {
			globerr = err
		}
	}

	return globerr
}

//Start the necessary servers to run StreamDB
func Kill(configuration *config.Configuration) error {
	log.Debugf("Killing subsystems")

	var globerr error
	sqlDatabaseType := configuration.DatabaseType

	switch sqlDatabaseType {
	case config.Postgres:
		if err := postgresInstance.Kill(); err != nil {
			globerr = err
		}
	case config.Sqlite:
		if err := sqliteInstance.Kill(); err != nil {
			globerr = err
		}
	default:
		globerr = ErrUnrecognizedDatabase
	}

	if err := gnatsdInstance.Kill(); err != nil {
		globerr = err
	}

	if err := redisInstance.Kill(); err != nil {
		globerr = err
	}

	return globerr
}
