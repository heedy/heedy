package dbmaker

import (
	"log"
	"path/filepath"
	"streamdb/util"
	"streamdb/config"
)


//Start the necessary servers to run StreamDB
func Start() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	log.Printf("Starting StreamDB from '%s'\n", streamdbDirectory)

	err = StartSqlDatabase()
	err = StartGnatsd()
	err = StartRedis()

	if err != nil {
		log.Print("Starting servers failed - shutting down\n", streamdbDirectory)
		StopSqlDatabase()
		StopGnatsd()
		StopRedis()
	} else {
		//The pid file doesn't actualyl contain a pid - it is more of a "run lock"
		util.Touch(filepath.Join(streamdbDirectory, "connectordb.pid"))
	}

	return err
}
