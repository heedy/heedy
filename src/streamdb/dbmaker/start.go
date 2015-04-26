package dbmaker

import (
	"errors"
	"log"
	"path/filepath"
)

//This error is thrown if a bad path is given as a database
var ErrNotDatabase = errors.New("The given path is not initialized as a database")

//Start the necessary servers to run StreamDB
func Start(streamdbDirectory, iface string, redisPort, gnatsdPort, sqlPort int, err error) error {
	if err == nil {
		if IsDirectory(streamdbDirectory) {
			streamdbDirectory, err = filepath.Abs(streamdbDirectory)
		} else {
			return ErrNotDatabase
		}

	}

	err = EnsureNotRunning(streamdbDirectory, err)
	if err != nil {
		return err
	}

	log.Printf("Starting StreamDB from '%s'\n", streamdbDirectory)

	err = StartSqlDatabase(streamdbDirectory, iface, sqlPort, err)
	err = StartGnatsd(streamdbDirectory, iface, gnatsdPort, err)
	err = StartRedis(streamdbDirectory, iface, redisPort, err)

	if err != nil {
		log.Print("Starting servers failed - shutting down\n", streamdbDirectory)
		StopSqlDatabase(streamdbDirectory, nil)
		StopGnatsd(streamdbDirectory, nil)
		StopRedis(streamdbDirectory, nil)
	} else {
		//The pid file doesn't actualyl contain a pid - it is more of a "run lock"
		Touch(filepath.Join(streamdbDirectory, "connectordb.pid"))
	}

	return err
}
