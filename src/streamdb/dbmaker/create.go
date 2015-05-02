package dbmaker

import (
	"errors"
	"log"
	"os"
	"streamdb/util"
	"streamdb/config"
	"streamdb/dbutil"
	"streamdb/users"
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

	log.Printf("Creating new StreamDB database at '%s'\n", streamdbDirectory)

	if err := os.MkdirAll(streamdbDirectory, FolderPermissions); err != nil {
		return err
	}

	if err := createSqlDatabase(config, username, password, email); err != nil {
		return err
	}

	if err := gnatsdInstance.Setup(); err != nil {
		return err
	}

	if err := redisInstance.Setup(); err != nil{
		return err
	}

	return nil
}


func createSqlDatabase(config *config.Configuration, username, password, email string) error {
	sqlDatabaseType := config.DatabaseType
	log.Printf("Creating sql database of type %s \n", sqlDatabaseType)

	switch sqlDatabaseType {
		case "postgres":
			if err := postgresInstance.Setup(); err != nil {
				return err
			}
		case "sqlite":
			if err := sqliteInstance.Setup(); err != nil {
				return err
			}
		default:
			return ErrUnrecognizedDatabase
	}

	log.Printf("Creating user %s (%s)\n", username, email)

	// Make the connection
	spath := config.GetDatabaseConnectionString()
	db, driver, err := dbutil.OpenSqlDatabase(spath)
	if err != nil {
		return err
	}

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
