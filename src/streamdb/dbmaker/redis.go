package dbmaker

import (
	"log"
	"streamdb/config"
)


//InitializeRedis sets up the configuration of redis
func InitializeRedis() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	log.Printf("Setting up Redis server\n")

	//Now copy the configuration file
	err = CopyConfig(streamdbDirectory, "redis.conf", err)

	return nil
}


//StartRedis runs the redis server
func StartRedis() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	host	:= config.GetConfiguration().RedisHost
	port    := config.GetConfiguration().RedisPort

	log.Printf("Starting Redis server on port %d\n", port)
	configfile, err := SetConfig(streamdbDirectory, "redis.conf",
		GenerateConfigReplacements(streamdbDirectory, "redis", host, port), err)

	err = RunDaemon(err, "redis-server", configfile)

	return WaitPort(host, port, err)
}


//StopRedis shuts down redis server
func StopRedis() error {
	streamdbDirectory, err := config.GetStreamdbDirectory()
	if err != nil {
		return err
	}

	log.Print("Stopping Redis server\n")

	return StopProcess(streamdbDirectory, "redis", err)
}
