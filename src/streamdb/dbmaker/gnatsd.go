package dbmaker

import (
	"log"
	"path/filepath"
	"streamdb/config"

	"github.com/kardianos/osext"
)


//InitializeGnatsd sets up the configuration of the gnatsd messaging daemon
func InitializeGnatsd() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	log.Printf("Setting up Gnatsd server\n")

	//Now copy the configuration file
	err = CopyConfig(streamdbDirectory, "gnatsd.conf", err)

	return nil
}

//StartGnatsd runs gnatsd
func StartGnatsd() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	host	:= config.GetConfiguration().GnatsdHost
	port    := config.GetConfiguration().GnatsdPort

	log.Printf("Starting gNATSd server on port %d\n", port)
	configfile, err := SetConfig(streamdbDirectory, "gnatsd.conf",
		GenerateConfigReplacements(streamdbDirectory, "gnatsd", host, port), err)

	execpath, err := osext.ExecutableFolder()

	//We assume gnatsd is distributed with our binaries in the dep folder
	gpath := filepath.Join(execpath, "dep/gnatsd")

	err = RunDaemon(err, gpath, "-c", configfile)

	return WaitPort(host, port, err)
}

//StopGnatsd stops the gnatsd server
func StopGnatsd() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	log.Print("Stopping gNATSd server\n")
	return StopProcess(streamdbDirectory, "gnatsd", err)

}
