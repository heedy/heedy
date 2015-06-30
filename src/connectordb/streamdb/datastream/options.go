package datastream

import (
	"github.com/apcera/nats"
	"gopkg.in/redis.v3"
)

//Options holds the configuration for connectordb
type Options struct {
	RedisOptions redis.Options
	BatchSize    int
	NatsOptions  nats.Options
}

//Default configuration options for datastream
var DefaultOptions = getDefaultOptions()

func getDefaultOptions() Options {
	var opt Options
	opt.NatsOptions = nats.DefaultOptions
	opt.RedisOptions = redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}
	opt.BatchSize = 250
	return opt
}
