package dbmaker

import (
	"log"
	"path/filepath"
	"streamdb/config"

	"github.com/kardianos/osext"
)



// A service representing the postgres database
type GnatsdService struct {
	ServiceHelper // We get stop, status, kill, and Name from this
	host string
	port int
	streamdbDirectory string
}


// Creates and returns a new postgres service in a pre-init state
// with default values loaded from config
func NewDefaultGnatsdService() *GnatsdService {
	return NewConfigGnatsdService(config.GetConfiguration())
}

func NewConfigGnatsdService(config *config.Configuration) *GnatsdService {
	host := config.GnatsdHost
	port := config.GnatsdPort
	dir  := config.StreamdbDirectory

	return NewGnatsdService(host, port, dir)
}

// Creates and returns a new postgres service in a pre-init state
func NewGnatsdService(host string, port int, streamdbDirectory string) *GnatsdService {
	var ps GnatsdService
	ps.host = host
	ps.port = port
	ps.streamdbDirectory = streamdbDirectory

	ps.InitServiceHelper(streamdbDirectory, "gnatsd")
	return &ps
}

//InitializeGnatsd sets up the configuration of the gnatsd messaging daemon
func (srv *GnatsdService) Setup() error {
	log.Printf("Setting up Gnatsd server\n")

	//Now copy the configuration file
	return CopyConfig(srv.streamdbDirectory, "gnatsd.conf", nil)
}

func (srv *GnatsdService) Init() error {
	log.Printf("Initializing Gnatsd\n")

	srv.Stat = StatusInit
	// Nothing to do here, may want to which/look for the executables in the
	// future and check the port is open
	return nil
}

//StartGnatsd runs gnatsd
func (srv *GnatsdService) Start() error {
	if srv.Stat == StatusRunning {
		return nil
	}
	if srv.Stat != StatusInit {
		log.Printf("Could not start gnatsd, status is %v\n", srv.Stat)
		return ErrNotInitialized
	}

	log.Printf("Starting gNATSd server on port %d\n", srv.port)

	configReplacements := GenerateConfigReplacements(srv.streamdbDirectory, "gnatsd", srv.host, srv.port)
	configfile, err := SetConfig(srv.streamdbDirectory, "gnatsd.conf", configReplacements , nil)

	execpath, err := osext.ExecutableFolder()

	//We assume gnatsd is distributed with our binaries in the dep folder
	gpath := filepath.Join(execpath, "dep/gnatsd")

	err = RunDaemon(err, gpath, "-c", configfile)

	err = WaitPort(srv.host, srv.port, err)
	if err != nil {
		srv.Stat = StatusRunning
	}

	return err
}


func (srv *GnatsdService) Stop() error {
	return srv.HelperStop()
}


func (srv *GnatsdService) Kill() error {
	return srv.HelperKill()
}
