package rediscache

import "gopkg.in/redis.v3"

type RedisBatchReader struct {
	client *redis.Client
}

func (r *RedisBatchReader) Close() {
	r.client.Close()
}
