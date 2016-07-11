package dbsetup

import (
	"config"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"util"

	"github.com/kardianos/osext"
)

//FrontendService is a service for running the ConnectorDB frontend
type FrontendService struct {
	BaseService
	c *config.Configuration
}

// The frontend doesn't need to be created.
func (s *FrontendService) Create() error {
	return nil
}

//Start starts the service
func (s *FrontendService) Start() error {
	// We just call our own executable with different options to start up frontend
	connectordb, err := osext.Executable()

	// This is the base run string
	flags := []string{"run", s.ServiceDirectory}

	// The greatest difficulty now is to figure out how to send command line options
	// from start. For now, we just use a hack: we manually set send ALL of the command line options.
	// This should be fixed at some point, but for now it works, so fukkit.
	flags = append(flags, "--loglevel", s.c.LogLevel)
	flags = append(flags, "--log", s.c.LogFile)

	var pid int
	pid, err = util.RunDaemon(err, connectordb, flags...)
	err = util.WaitPort(s.c.Hostname, int(s.c.Port), err)

	if err == nil {
		s.Stat = StatusRunning
	} else {
		s.Stat = StatusError
	}
	if err != nil {
		return err
	}

	// Finally, we write a pid file for frontend
	return ioutil.WriteFile(filepath.Join(s.ServiceDirectory, "frontend.pid"), []byte(strconv.Itoa(pid)), 0666)

}

//NewFrontendService creates a new service for the ConnectorDB frontend
func NewFrontendService(serviceDirectory string, c *config.Configuration) *FrontendService {
	return &FrontendService{BaseService{serviceDirectory, "frontend", StatusNone, nil}, c}
}
