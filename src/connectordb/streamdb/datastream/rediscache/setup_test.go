package rediscache

import (
	"os"
	"testing"

	"connectordb/config"

	log "github.com/Sirupsen/logrus"
)

var (
	rc  *RedisConnection
	err error
)

func TestMain(m *testing.M) {

	rc, err = NewRedisConnection(&config.DefaultOptions.RedisOptions)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	rc.Close()

	rc, err = NewRedisConnection(&config.DefaultOptions.RedisOptions)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	res := m.Run()

	rc.Close()
	os.Exit(res)
}
