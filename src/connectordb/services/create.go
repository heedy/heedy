package services

import (
	"connectordb/config"
	"connectordb/streamdb/dbutil"
	"connectordb/streamdb/users"
	"connectordb/streamdb/util"
	"errors"
	"os"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrDirectoryExists is thrown if try to init over existing directory
	ErrDirectoryExists = errors.New("Cannot initialize database in an existing directory")
	//ErrUnrecognizedDatabase is thrown when the database type is not recognized
	ErrUnrecognizedDatabase = errors.New("Unrecognized sql database type")
)

//Create a streamdb instance
func Create(config *config.Configuration, username, password, email string) error {
	streamdbDirectory := config.StreamdbDirectory

	if util.PathExists(streamdbDirectory) {
		return ErrDirectoryExists
	}

	log.Printf("Creating new StreamDB database at '%s'", streamdbDirectory)

	if err := os.MkdirAll(streamdbDirectory, FolderPermissions); err != nil {
		return err
	}

	if err := createSqlDatabase(config, username, password, email); err != nil {
		return err
	}

	if err := gnatsdInstance.Setup(); err != nil {
		return err
	}

	if err := redisInstance.Setup(); err != nil {
		return err
	}

	return nil
}

func createSqlDatabase(configuration *config.Configuration, username, password, email string) error {
	sqlDatabaseType := configuration.DatabaseType
	log.Printf("Creating sql database of type %s", sqlDatabaseType)

	switch sqlDatabaseType {
	case config.Postgres:
		if err := postgresInstance.Setup(); err != nil {
			return err
		}
	case config.Sqlite:
		if err := sqliteInstance.Setup(); err != nil {
			return err
		}
	default:
		return ErrUnrecognizedDatabase
	}

	log.Printf("Creating user %s (%s)", username, email)

	// Make the connection
	spath := configuration.GetDatabaseConnectionString()
	db, driver, err := dbutil.OpenSqlDatabase(spath)
	if err != nil {
		return err
	}
	defer db.Close()

	var udb users.UserDatabase
	udb.InitUserDatabase(db, string(driver))
	err = udb.CreateUser(username, email, password)
	if err != nil {
		return err
	}

	usr, err := udb.ReadUserByName(username)
	if err != nil {
		return err
	}

	usr.Admin = true
	return udb.UpdateUser(usr)
}
