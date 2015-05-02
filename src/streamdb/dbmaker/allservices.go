package dbmaker

import (
	"streamdb/config"
	"streamdb/util"
	"path/filepath"
	"os"
	"errors"
	"log"
)

var(
	sqliteInstance 		*SqliteService
	postgresInstance 	*PostgresService
	gnatsdInstance 		*GnatsdService
	redisInstance 		*RedisService

	ErrNotInitialized = errors.New("Module not yet initialized")
	ErrAlreadyInitialized = errors.New("All subsystems have allready been initialized")
	doneInit = false
)


func initSqlDatabase(config *config.Configuration) error {
	if doneInit {
		return ErrAlreadyInitialized
	}

	sqliteInstance = NewConfigSqliteSerivce(config)
	postgresInstance = NewConfigPostgresService(config)

	sqlDatabaseType := config.DatabaseType

	switch sqlDatabaseType {
		case "postgres":
			if err := postgresInstance.Init(); err != nil {
				return err
			}
		case "sqlite":
			if err := sqliteInstance.Init(); err != nil {
				return err
			}
		default:
			return ErrUnrecognizedDatabase
	}

	return nil
}

func Init(config *config.Configuration) error {
	if doneInit {
		return ErrAlreadyInitialized
	}

	log.Printf("Initializing subsystems\n")
	gnatsdInstance = NewConfigGnatsdService(config)
	redisInstance = NewConfigRedisService(config)

	if err := initSqlDatabase(config); err != nil {
		return err
	}

	if err := gnatsdInstance.Init(); err != nil {
		return err
	}

	if err := redisInstance.Init(); err != nil{
		return err
	}
	log.Printf("Finished initializing subsystems\n")

	doneInit = true
	return nil
}

func startSqlDatabase(config *config.Configuration) error {
	sqlDatabaseType := config.DatabaseType

	switch sqlDatabaseType {
		case "postgres":
			if err := postgresInstance.Start(); err != nil {
				return err
			}
		case "sqlite":
			if err := sqliteInstance.Start(); err != nil {
				return err
			}
		default:
			return ErrUnrecognizedDatabase
	}

	return nil
}


//Start the necessary servers to run StreamDB
func Start(config *config.Configuration) error {
	log.Printf("Starting subsystems\n")

	os.Chdir(config.StreamdbDirectory)

	if err := startSqlDatabase(config); err != nil {
		return err
	}

	if err := gnatsdInstance.Start(); err != nil {
		return err
	}

	if err := redisInstance.Start(); err != nil{
		return err
	}

	util.Touch(filepath.Join(config.StreamdbDirectory, "connectordb.pid"))

	return nil
}


func stopSqlDatabase(config *config.Configuration) error{
	sqlDatabaseType := config.DatabaseType

	switch sqlDatabaseType {
		case "postgres":
			return postgresInstance.Stop()
		case "sqlite":
			return sqliteInstance.Stop()
	}
	return ErrUnrecognizedDatabase
}


//Start the necessary servers to run StreamDB
func Stop(config *config.Configuration) error {
	log.Printf("Stopping subsystems\n")

	var globerr error
	if err := stopSqlDatabase(config); err != nil {
		globerr = err
	}

	if err := gnatsdInstance.Stop(); err != nil {
		globerr = err
	}

	if err := redisInstance.Stop(); err != nil{
		globerr = err
	}

	pidpath := filepath.Join(config.StreamdbDirectory, "connectordb.pid")
	if util.PathExists(pidpath) {
		if err := os.Remove(pidpath); err != nil {
			globerr = err
		}
	}

	return globerr
}


//Start the necessary servers to run StreamDB
func Kill(config *config.Configuration) error {
	log.Printf("Killing subsystems\n")

	var globerr error
	sqlDatabaseType := config.DatabaseType

	switch sqlDatabaseType {
		case "postgres":
			if err := postgresInstance.Kill(); err != nil {
				globerr = err
			}
		case "sqlite":
			if err := sqliteInstance.Kill(); err != nil {
				globerr = err
			}
		default:
			globerr = ErrUnrecognizedDatabase
	}

	if err := gnatsdInstance.Kill(); err != nil {
		globerr = err
	}

	if err := redisInstance.Kill(); err != nil{
		globerr = err
	}

	return globerr
}
