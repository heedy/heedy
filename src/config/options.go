/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"fmt"

	"github.com/nats-io/nats"
	"gopkg.in/redis.v3"
)

// Options are the struct which gives ConnectorDB core the necessary information
// to connect to the underlying servers.
type Options struct {
	RedisOptions redis.Options
	NatsOptions  nats.Options

	SqlConnectionString string

	UserCacheSize   int64
	DeviceCacheSize int64
	StreamCacheSize int64
	CacheEnabled    bool

	BatchSize int // BatchSize is the number of datapoints per batch of data in a stream
	ChunkSize int // ChunkSize is the number of batches to queue up before writing to storage
}

func (o *Options) String() string {
	return fmt.Sprintf(`ConnectorDB:
Batch Size: %v
Chunk Size: %v

Redis: %v (%v)
Nats:  %v
Sql:   %v
`, o.BatchSize, o.ChunkSize, o.RedisOptions.Addr, o.RedisOptions.Password, o.NatsOptions.Url, o.SqlConnectionString)
}

//Options generates the ConnectorDB options based upon the given configuration
func (c *Configuration) Options() *Options {
	var opt Options

	opt.NatsOptions = nats.DefaultOptions
	opt.NatsOptions.Url = c.Nats.GetNatsConnectionString()

	opt.RedisOptions = redis.Options{
		Addr:     c.Redis.GetRedisConnectionString(),
		Password: c.Redis.Password,
		DB:       0,
	}

	opt.SqlConnectionString = c.GetSqlConnectionString()

	opt.BatchSize = c.BatchSize
	opt.ChunkSize = c.ChunkSize

	opt.CacheEnabled = c.UseCache
	opt.DeviceCacheSize = c.DeviceCacheSize
	opt.UserCacheSize = c.UserCacheSize
	opt.StreamCacheSize = c.StreamCacheSize

	return &opt
}
