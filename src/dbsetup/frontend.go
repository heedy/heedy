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
	o *Options
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

	// Set up the frontend's flags
	if len(s.o.FrontendFlags) > 0 {
		flags = append(flags, s.o.FrontendFlags...)
	}

	var pid int
	pid, err = util.RunDaemon(err, connectordb, flags...)

	// The port might have been modified by flag. Check if that is the case
	port := s.c.Port
	if s.o.FrontendPort != 0 {
		port = s.o.FrontendPort
	}
	// Windows needs to know that we're on localhost
	err = util.WaitPort("localhost", int(port), err)

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
func NewFrontendService(serviceDirectory string, c *config.Configuration, o *Options) *FrontendService {
	return &FrontendService{BaseService{serviceDirectory, "frontend", StatusNone, nil}, c, o}
}
