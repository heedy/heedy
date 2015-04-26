package dbmaker

import (
	"log"
	"os"
	"path/filepath"
)

//Stop shuts down all services running in the given directory
func Stop(streamdbDirectory string, err error) error {
	if err == nil {
		if IsDirectory(streamdbDirectory) {
			streamdbDirectory, err = filepath.Abs(streamdbDirectory)
		} else {
			return ErrNotDatabase
		}

	}

	log.Printf("Stopping database in '%s'\n", streamdbDirectory)

	err = StopGnatsd(streamdbDirectory, nil)
	if err != nil {
		log.Printf("Error Stopping Gnatsd: %v", err)
	}
	err = StopRedis(streamdbDirectory, nil)
	if err != nil {
		log.Printf("Error Stopping Redis: %v", err)
	}
	err = StopPostgres(streamdbDirectory, nil)
	if err != nil {
		log.Printf("Error Stopping Postgres: %v", err)
	}

	err2 := os.Remove(filepath.Join(streamdbDirectory, "connectordb.pid"))
	if err2 != nil {
		err = err2
	}
	return err
}
