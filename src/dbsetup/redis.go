/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"strconv"
	"util"

	log "github.com/Sirupsen/logrus"
)

//RedisService is a service for running Redis
type RedisService struct {
	BaseService
}

//Start starts the service
func (s *RedisService) Start() error {
	configfile, err := s.start()
	if err != nil {
		return err
	}

	err = util.RunDaemon(err, "redis-server", configfile)
	err = util.WaitPort(s.S.Hostname, int(s.S.Port), err)

	if err == nil {
		s.Stat = StatusRunning
	} else {
		s.Stat = StatusError
	}

	return err
}

//Stop shuts down the redis server
func (s *RedisService) Stop() error {
	if s == nil {
		return nil
	}
	log.Print("Stopping redis...")
	portString := strconv.Itoa(int(s.S.Port))

	return util.RunCommand(nil, "redis-cli", "-p", portString, "-a", s.S.Password, "shutdown")
}

//NewRedisService creates a new service for Redis
func NewRedisService(serviceDirectory string, s *config.Service) *RedisService {
	return &RedisService{BaseService{serviceDirectory, "redis", StatusNone, s}}
}
