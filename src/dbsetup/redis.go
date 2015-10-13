package dbsetup

import (
	"strconv"
	"util"

	log "github.com/Sirupsen/logrus"
)

//RedisService is a service for running Redis
type RedisService struct {
	BaseService
	Host string
	Port int
}

//Create prepares redis
func (s *RedisService) Create() error {
	log.Infof("Setting up Redis server")
	//Redis does not need major setup - we just copy the configuration file and we're ready to go
	return CopyConfig(s.ServiceDirectory, "redis.conf", nil)
}

//Start starts the service
func (s *RedisService) Start() error {
	if s.Status() == StatusRunning {
		return nil
	}

	log.Infof("Staring redis on port %d", s.Port)

	configReplacements := GenerateConfigReplacements(s.ServiceDirectory, "redis", s.Host, s.Port)
	configfile, err := SetConfig(s.ServiceDirectory, "redis.conf", configReplacements, nil)
	if err != nil {
		return err
	}

	err = util.RunDaemon(err, "redis-server", configfile)
	err = util.WaitPort(s.Host, s.Port, err)

	if err == nil {
		s.Stat = StatusRunning
	} else {
		s.Stat = StatusError
	}

	return err
}

//Stop shuts down the redis server
func (s *RedisService) Stop() error {
	log.Print("Stopping redis...")
	portString := strconv.Itoa(s.Port)

	return util.RunCommand(nil, "redis-cli", "-p", portString, "shutdown")
}

//NewRedisService creates a new service for Redis
func NewRedisService(serviceDirectory, host string, port int) *RedisService {
	return &RedisService{BaseService{serviceDirectory, "redis", StatusNone}, host, port}
}
