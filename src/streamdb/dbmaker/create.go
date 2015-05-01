package dbmaker

import (
	"errors"
	"log"
	"os"
	"streamdb/util"
	"streamdb/config"
)

var (
	//ErrDirectoryExists is thrown if try to init over existing directory
	ErrDirectoryExists = errors.New("Cannot initialize database in an existing directory")
	//ErrUnrecognizedDatabase is thrown when the database type is not recognized
	ErrUnrecognizedDatabase = errors.New("Unrecognized sql database type")
)

/*
This package contains the necessary tools to create and initialize a full streamdb database
*/

//Create initializes a full StreamDB database
func Create() error {

	sqlDatabaseType := config.GetConfiguration().DatabaseType
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	if util.PathExists(streamdbDirectory) {
		return ErrDirectoryExists
	}

	log.Printf("Creating new StreamDB database at '%s'\n", streamdbDirectory)

	err = os.MkdirAll(streamdbDirectory, FolderPermissions)

	switch sqlDatabaseType {
		case "postgres":
			if err := InitializePostgres(); err != nil {
				return err
			}
		case "sqlite":
			if err := InitializeSqlite(); err != nil {
				return err
			}
		default:
			os.RemoveAll(streamdbDirectory)
			return ErrUnrecognizedDatabase
	}

	if err := InitializeGnatsd(); err != nil{
		return err
	}
	err = InitializeRedis()

	return err
}
