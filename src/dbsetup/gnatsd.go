/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"path/filepath"
	"util"

	"github.com/kardianos/osext"
)

//GnatsdService is a service for running Redis
type GnatsdService struct {
	BaseService
}

//Start starts the service
func (s *GnatsdService) Start() error {
	configfile, err := s.start()
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
	err = util.WaitPort(s.S.Hostname, int(s.S.Port), err)

	if err == nil {
		s.Stat = StatusRunning
	} else {
		s.Stat = StatusError
	}

	return err
}

//NewGnatsdService creates a new service for gNatsd
func NewGnatsdService(serviceDirectory string, s *config.Service) *GnatsdService {
	return &GnatsdService{BaseService{serviceDirectory, "gnatsd", StatusNone, s}}
}
