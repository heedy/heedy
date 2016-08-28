/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"strconv"
	"time"
	"util"

	redis "gopkg.in/redis.v4"

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

	_, err = util.RunDaemon(err, GetExecutablePath("redis-server"), configfile)
	err = util.WaitPort(s.S.Hostname, int(s.S.Port), err)

	if err == nil {
		s.Stat = StatusRunning
	} else {
		s.Stat = StatusError
		return err
	}

	// Now wait until redis finished loading dataset into memory
	rclient := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + strconv.Itoa(int(s.S.Port)),
		Password: s.S.Password,
		DB:       0,
	})

	_, err = rclient.Ping().Result()
	if err != nil && err.Error() == "LOADING Redis is loading the dataset in memory" {
		log.Debug("Waiting for Redis to load dataset...")
		for err != nil && err.Error() == "LOADING Redis is loading the dataset in memory" {
			time.Sleep(300 * time.Millisecond)
			_, err = rclient.Ping().Result()
		}
		if err == nil {
			log.Debug("Redis finished loading dataset.")
		}
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

	return util.RunCommand(nil, GetExecutablePath("redis-cli"), "-p", portString, "-a", s.S.Password, "shutdown")
}

//NewRedisService creates a new service for Redis
func NewRedisService(serviceDirectory string, s *config.Service) *RedisService {
	return &RedisService{BaseService{serviceDirectory, "redis", StatusNone, s}}
}
