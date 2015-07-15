package config

import (
	"fmt"

	"github.com/nats-io/nats"
	"gopkg.in/redis.v3"
)

//DefaultOptions are the options that work with all default connection settings
var DefaultOptions = NewOptions()

//Options are the struct which gives ConnectorDB core the necessary information
//to connect to the underlying servers.
type Options struct {
	RedisOptions redis.Options
	NatsOptions  nats.Options

	SqlConnectionType   string
	SqlConnectionString string

	BatchSize int //BatchSize is the number of datapoints per batch of data in a stream
	ChunkSize int //ChunkSize is the number of batches to queue up before writing to storage
}

func (o *Options) String() string {
	return fmt.Sprintf(`StreamDB Options
Sql Database Type: %s
Sql Connection String: %s

Batch Size: %v
Chunk Size: %v

Redis Options: %v
Nats Options: %v
`, o.SqlConnectionType, o.SqlConnectionString, o.BatchSize, o.ChunkSize, o.RedisOptions, o.NatsOptions)
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

	opt.SqlConnectionType = Postgres
	opt.SqlConnectionString = "sslmode=disable dbname=connectordb port=52592"

	opt.BatchSize = 250
	opt.ChunkSize = 1

	return &opt
}
