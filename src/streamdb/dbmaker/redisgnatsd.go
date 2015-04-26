package dbmaker

import (
	"log"
	"path/filepath"

	"github.com/kardianos/osext"
)

//InitializeGnatsd sets up the configuration of the gnatsd messaging daemon
func InitializeGnatsd(streamdbDirectory string, err error) error {
	if err != nil {
		return err
	}
	log.Printf("Setting up Gnatsd server\n")

	//Now copy the configuration file
	err = CopyConfig(streamdbDirectory, "redis.conf", err)

	return nil
}

//InitializeRedis sets up the configuration of redis
func InitializeRedis(streamdbDirectory string, err error) error {
	if err != nil {
		return err
	}

	log.Printf("Setting up Redis server\n")

	//Now copy the configuration file
	err = CopyConfig(streamdbDirectory, "gnatsd.conf", err)

	return nil
}

//StartGnatsd runs gnatsd
func StartGnatsd(streamdbDirectory, iface string, port int, err error) error {
	if err != nil {
		return err
	}

	log.Printf("Starting gNATSd server on port %d\n", port)
	configfile, err := SetConfig(streamdbDirectory, "gnatsd.conf",
		GenerateConfigReplacements(streamdbDirectory, "gnatsd", iface, port), err)

	execpath, err := osext.ExecutableFolder()

	//We assume gnatsd is distributed with our binaries in the dep folder
	gpath := filepath.Join(execpath, "dep/gnatsd")

	err = RunDaemon(err, gpath, "-c", configfile)

	return WaitPort(iface, port, err)
}

//StartRedis runs the redis server
func StartRedis(streamdbDirectory, iface string, port int, err error) error {
	if err != nil {
		return err
	}

	log.Printf("Starting Redis server on port %d\n", port)
	configfile, err := SetConfig(streamdbDirectory, "redis.conf",
		GenerateConfigReplacements(streamdbDirectory, "redis", iface, port), err)

	err = RunDaemon(err, "redis-server", configfile)

	return WaitPort(iface, port, err)
}

//StopGnatsd stops the gnatsd server
func StopGnatsd(streamdbDirectory string, err error) error {
	if err != nil {
		return err
	}

	log.Print("Stopping gNATSd server\n")
	return StopProcess(streamdbDirectory, "gnatsd", err)

}

//StopRedis shuts down redis server
func StopRedis(streamdbDirectory string, err error) error {
	if err != nil {
		return err
	}

	log.Print("Stopping Redis server\n")

	return StopProcess(streamdbDirectory, "redis", err)
}
