package dbsetup

import (
	"path/filepath"
	"util"

	log "github.com/Sirupsen/logrus"
	"github.com/kardianos/osext"
)

//GnatsdService is a service for running Redis
type GnatsdService struct {
	BaseService
	Host string
	Port int
}

//Create prepares redis
func (s *GnatsdService) Create() error {
	log.Infof("Setting up Gnatsd server")
	//Redis does not need major setup - we just copy the configuration file and we're ready to go
	return CopyConfig(s.ServiceDirectory, "gnatsd.conf", nil)
}

//Start starts the service
func (s *GnatsdService) Start() error {
	if s.Status() == StatusRunning {
		return nil
	}

	log.Infof("Staring gNATSd on port %d", s.Port)

	configReplacements := GenerateConfigReplacements(s.ServiceDirectory, "gnatsd", s.Host, s.Port)
	configfile, err := SetConfig(s.ServiceDirectory, "gnatsd.conf", configReplacements, nil)
	if err != nil {
		return err
	}

	execpath, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}

	//We assume gnatsd is distributed with our binaries in the dep folder
	gpath := filepath.Join(execpath, "dep/gnatsd")

	err = util.RunDaemon(err, gpath, "-c", configfile)
	err = util.WaitPort(s.Host, s.Port, err)

	if err == nil {
		s.Stat = StatusRunning
	} else {
		s.Stat = StatusError
	}

	return err
}

//NewGnatsdService creates a new service for gNatsd
func NewGnatsdService(serviceDirectory, host string, port int) *GnatsdService {
	return &GnatsdService{BaseService{serviceDirectory, "gnatsd", StatusNone}, host, port}
}
