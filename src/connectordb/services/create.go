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
	log.Debugf("Creating sql database")

	if err := postgresInstance.Setup(); err != nil {
		postgresInstance.Stop()
		return err
	}

	log.Infof("Creating user %s (%s)", username, email)

	// Make the connection
	spath := configuration.GetDatabaseConnectionString()
	db, driver, err := dbutil.OpenSqlDatabase(spath)
	if err != nil {
		postgresInstance.Stop()
		return err
	}
	defer db.Close()

	udb := users.NewUserDatabase(db, driver, false)
	err = udb.CreateUser(username, email, password)
	if err != nil {
		postgresInstance.Stop()
		return err
	}

	usr, err := udb.ReadUserByName(username)
	if err != nil {
		postgresInstance.Stop()
		return err
	}

	usr.Admin = true
	err = udb.UpdateUser(usr)
	if err != nil {
		postgresInstance.Stop()
	}
	return err
}
