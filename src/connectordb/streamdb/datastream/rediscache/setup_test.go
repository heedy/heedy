package rediscache

import (
	"os"
	"testing"

	"gopkg.in/redis.v3"

	log "github.com/Sirupsen/logrus"
)

var (
	rc  *RedisConnection
	err error
)

func TestMain(m *testing.M) {

	rc, err = NewRedisConnection(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	rc.Close()

	rc, err = NewRedisConnection(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	res := m.Run()

	rc.Close()
	os.Exit(res)
}
