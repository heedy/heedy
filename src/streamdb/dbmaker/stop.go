package dbmaker

import (
	"log"
	"os"
	"path/filepath"
	"streamdb/util"
	"streamdb/config"
)

//Stop shuts down all services running in the given directory
func Stop() error {

	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	if util.IsDirectory(streamdbDirectory) {
		streamdbDirectory, err = filepath.Abs(streamdbDirectory)
	} else {
		return util.ErrNotDatabase
	}

	log.Printf("Stopping database in '%s'\n", streamdbDirectory)

	err = StopGnatsd()
	if err != nil {
		log.Printf("Error Stopping Gnatsd: %v", err)
	}
	err = StopRedis()
	if err != nil {
		log.Printf("Error Stopping Redis: %v", err)
	}
	err = StopPostgres()
	if err != nil {
		log.Printf("Error Stopping Postgres: %v", err)
	}

	err2 := os.Remove(filepath.Join(streamdbDirectory, "connectordb.pid"))
	if err2 != nil {
		err = err2
	}
	return err
}
