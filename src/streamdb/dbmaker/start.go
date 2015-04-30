package dbmaker

import (
	"log"
	"path/filepath"
	"streamdb/util"
)


//Start the necessary servers to run StreamDB
func Start(streamdbDirectory, iface string, redisPort, gnatsdPort, sqlPort int, err error) error {
	if err != nil {
		return err
	}

	streamdbDirectory, err = util.ProcessConnectordbDirectory(streamdbDirectory)
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
		util.Touch(filepath.Join(streamdbDirectory, "connectordb.pid"))
	}

	return err
}
