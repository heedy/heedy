package dbmaker

import (
	"errors"
	"log"
	"os"
	"path/filepath"
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
func Create(streamdbDirectory, sqlDatabaseType string, err error) error {

	//Make sure we are using an absolute path
	streamdbDirectory, err = filepath.Abs(streamdbDirectory)
	if err != nil {
		return err
	}
	log.Printf("Creating new StreamDB database at '%s'\n", streamdbDirectory)

	if PathExists(streamdbDirectory) {
		return ErrDirectoryExists
	}

	err = os.MkdirAll(streamdbDirectory, FolderPermissions)

	switch sqlDatabaseType {
	case "postgres":
		err = InitializePostgres(streamdbDirectory, err)
	case "sqlite":
		err = InitializeSqlite(streamdbDirectory, err)
	default:
		os.RemoveAll(streamdbDirectory)
		return ErrUnrecognizedDatabase
	}

	err = InitializeGnatsd(streamdbDirectory, err)
	err = InitializeRedis(streamdbDirectory, err)

	return err
}
