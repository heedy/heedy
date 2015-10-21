package config

import (
	"fmt"

	"github.com/nats-io/nats"
	"gopkg.in/redis.v3"
)

//Options are the struct which gives ConnectorDB core the necessary information
//to connect to the underlying servers.
type Options struct {
	RedisOptions redis.Options
	NatsOptions  nats.Options

	SqlConnectionString string

	BatchSize int //BatchSize is the number of datapoints per batch of data in a stream
	ChunkSize int //ChunkSize is the number of batches to queue up before writing to storage
}

func (o *Options) String() string {
	return fmt.Sprintf(`ConnectorDB Options
Sql Connection String: %s

Batch Size: %v
Chunk Size: %v

Redis Address: %v (%v)
Nats Address: %v
`, o.SqlConnectionString, o.BatchSize, o.ChunkSize, o.RedisOptions.Addr, o.RedisOptions.Password, o.NatsOptions.Url)
}

//NewOptions returns new options set to default values.
func NewOptions() *Options {
	var opt Options
	opt.NatsOptions = nats.DefaultOptions
	opt.NatsOptions.Url = nats.DefaultURL

	opt.RedisOptions = redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	opt.SqlConnectionString = "sslmode=disable dbname=connectordb port=52592"

	opt.BatchSize = 250
	opt.ChunkSize = 1

	return &opt
}
