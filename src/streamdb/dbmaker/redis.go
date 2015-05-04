package dbmaker

import (
	"log"
	"streamdb/config"
	"strconv"
)


// A service representing the postgres database
type RedisService struct{
	ServiceHelper // We get stop, status, kill, and Name from this
	host string
	port int
	streamdbDirectory string
}


// Creates and returns a new postgres service in a pre-init state
// with default values loaded from config
func NewDefaultRedisService() *RedisService {
	return NewConfigRedisService(config.GetConfiguration())
}

// Creates a redis service from a given configuration
func NewConfigRedisService(config *config.Configuration) *RedisService  {
	host := config.RedisHost
	port := config.RedisPort
	dir  := config.StreamdbDirectory

	return NewRedisService(host, port, dir)
}

// Creates and returns a new postgres service in a pre-init state
func NewRedisService(host string, port int, streamdbDirectory string) *RedisService {
	var ps RedisService
	ps.host = host
	ps.port = port
	ps.streamdbDirectory = streamdbDirectory

	ps.InitServiceHelper(streamdbDirectory, "redis")
	return &ps
}

//InitializeRedis sets up the configuration of redis
func (srv *RedisService) Setup() error {
	log.Printf("Setting up Redis server\n")

	//Now copy the configuration file
	return CopyConfig(srv.streamdbDirectory, "redis.conf", nil)
}


func (srv *RedisService) Init() error {
	log.Printf("Initializing redis\n")
	srv.Stat = StatusInit
	// Nothing to do here, may want to which/look for the executables in the
	// future and check the port is open
	return nil
}

//StartRedis runs the redis server
func (srv *RedisService) Start() error {
	if srv.Stat == StatusRunning {
		return nil
	}
	if srv.Stat != StatusInit {
		log.Printf("Could not start redis, status is %v\n", srv.Stat)
		return ErrNotInitialized
	}

	log.Printf("Starting Redis server on port %d\n", srv.port)

	configReplacements := GenerateConfigReplacements(srv.streamdbDirectory, "redis", srv.host, srv.port)
	configfile, err := SetConfig(srv.streamdbDirectory, "redis.conf", configReplacements, nil)
	if err != nil {
		return err
	}

	log.Println(configfile)

	err = RunDaemon(err, "redis-server", configfile)
	err = WaitPort(srv.host, srv.port, err)

	if err != nil {
		srv.Stat = StatusRunning
	}

	return err
}


func (srv *RedisService) Stop() error {
	portString := strconv.Itoa(srv.port)

	return RunCommand(nil, "redis-cli", "-p", portString, "shutdown")
	//return srv.HelperStop()
}


func (srv *RedisService) Kill() error {
	return srv.HelperKill()
}
