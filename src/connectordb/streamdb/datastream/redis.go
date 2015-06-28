package datastream

import (
	"errors"
	"strconv"
	"strings"

	"gopkg.in/redis.v3"
)

/*
The structure in Redis is as follows:

Each stream has one list and one hset.
Datapoints are inserted in chunks which can have anywhere from 1 to chunksize datapoints,
all marshalled into bytes using messagepack
Suppose that we have stream "mystream"

	'{mystream}::s' = [(chunk),(chunk),(chunk)]

	'{mystream}:m' = {
		'endtime:' : the most recent timestamp of inserted data
		'starttime:' : the first timestamp of data in redis
		'length:' : the total number of datapoints in the stream (overall)
	}

Notice that most elements end in : (or have extra :).
This is because streams can have substreams. A stream 'mystream' with a downlink substream 'd'
would look like this:

	'{mystream}::s' = [(chunk),(chunk),(chunk)]
	'{mystream}:d:s' = [(chunk),(chunk),(chunk)]

	'{mystream}:m' = {
		'endtime:' : the most recent timestamp of inserted data
		'starttime:' : the first timestamp of data in redis
		'length:' : the total number of datapoints in the stream (overall)

		'endtime:d' : the most recent timestamp of inserted data in 'd'
		'starttime:d' : the first timestamp of data in redis for substream
		'length:d' : the total number of datapoints in the substream 'd'
	}

Lastly, notice the {} around mystream - this is for redis cluster hashing.
It allows all keys relevant to a stream to be on the same redis instance,
which is exploited heavily in the scripts.

*/

const (
	//RedisNilString is the string error that is returned if redis gives nil
	RedisNilString = "redis: nil"

	//The insert script does the following:
	//It is given 2 keys:
	//	stream key - the key where a list of chunks has been inserted
	//	metadata key - the key where the stream's metadata is stored
	//Of the arguments, it is given:
	//	substream - the name of the substream
	//	starttime - the start time of the datapoints
	//	endtime - the end time of the datapoints
	//	restamp - whether to restamp datapoints if inconsistent timestamps
	//	... array of the chunks to be inserted ...
	insertScript = `
		-- Make sure that the timestamps are increasing
		local stream_endtime = tonumber(redis.call('hget',KEYS[2], 'endtime:' .. ARGV[1]))
		if (stream_endtime~=nil and stream_endtime > tonumber(ARGV[2])) then
			if (ARGV[4] == '0') then
				return {["err"]="Greater timestamp already exists for the stream. Insert Failed."}
			end

			-- Restamp is ON. Go backwards from the end of the array while the timestamp
			-- is less than stream_endtime, and repack the msgpack with stream_endtime

			-- First, we work around an annoyance in lua's implementation of msgpack
			-- it encodes 5.0 as int - meaning that floats can lose floatiness
			if (math.floor(stream_endtime)==stream_endtime) then
				stream_endtime = stream_endtime + 0.00001
			end

			for i=#ARGV,5,-1 do
				local val = cmsgpack.unpack(ARGV[i])
				if (val['t'] > stream_endtime) then
					break
				end
				val['t'] = stream_endtime
				ARGV[i] = cmsgpack.pack(val)
			end

			-- The endtime might also be messed up - set it if that's the case
			if (tonumber(ARGV[3]) < stream_endtime) then
				ARGV[3] = stream_endtime
			end
		end
		-- Set the end time
		redis.call('hset',KEYS[2], 'endtime:' .. ARGV[1], ARGV[3])
		-- Set the total stream length
		redis.call('hincrby',KEYS[2], 'length:' .. ARGV[1], #ARGV - 4)

		-- Insert the datapoints into the stream - redis lua has some weird stuff about the maximum
		-- number of arguments to a function - we avoid this by manually splitting
		for i=5,#ARGV,5000 do
			redis.call('rpush',KEYS[1], unpack(ARGV,i,math.min(i+4999,#ARGV)))
		end
	`

	//The subdelete script deletes a given substream.
	//Given 2 keys:
	//	the stream key
	//	metadata key
	//In arguments it is given:
	//	the substream to delete
	subdeleteScript = `
		redis.call('del',KEYS[1])
		redis.call('hdel',KEYS[2],'endtime:' .. ARGV[1], 'length:' .. ARGV[1], 'starttime:' .. ARGV[1])
	`
)

var (
	//ErrTimestamp is returned when trying to insert old timestamps
	ErrTimestamp = errors.New("Greater timestamp already exists for the stream. Insert Failed.")
)

//RedisConnection is the connection to redis server
type RedisConnection struct {
	redis *redis.Client

	//The server side scripts which speed up certain operations
	insertScript    *redis.Script
	subdeleteScript *redis.Script
}

func wrapNil(err error) error {
	if err != nil && err.Error() == RedisNilString {
		return nil
	}
	return err
}

//Close the redis cluster connection
func (rc *RedisConnection) Close() {
	rc.redis.Close()
}

//NewRedisConnection creates a new connection to a redis server
func NewRedisConnection(opt *Options) (rc *RedisConnection, err error) {
	rclient := redis.NewClient(&opt.RedisOptions)

	_, err = rclient.Ping().Result()

	return &RedisConnection{
		redis:           rclient,
		insertScript:    redis.NewScript(insertScript),
		subdeleteScript: redis.NewScript(subdeleteScript),
	}, err
}

//Get returns all the datapoints cached associated with the given stream/substream
func (rc *RedisConnection) Get(stream, substream string) (dpa DatapointArray, err error) {
	sa, err := rc.redis.LRange(stream+":"+substream+":s", 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return DatapointArrayFromDataStrings(sa)

}

//Insert datapoint array
func (rc *RedisConnection) Insert(stream, substream string, dpa DatapointArray, restamp bool) (err error) {
	//remember the number of args here
	args := make([]string, 4+len(dpa))

	args[0] = substream
	args[1] = strconv.FormatFloat(dpa[0].Timestamp, 'G', -1, 64)
	args[2] = strconv.FormatFloat(dpa[len(dpa)-1].Timestamp, 'G', -1, 64)
	if restamp {
		args[3] = "1"
	} else {
		args[3] = "0"
	}

	for i := range dpa {
		b, err := dpa[i].Bytes()
		if err != nil {
			return err
		}
		args[i+4] = string(b)
	}

	return wrapNil(rc.insertScript.Run(rc.redis, []string{stream + ":" + substream + ":s", stream + ":m"},
		args).Err())
}

//StreamLength returns the stream's length
func (rc *RedisConnection) StreamLength(stream, substream string) (int64, error) {
	sc := rc.redis.HGet(stream+":m", "length:"+substream)

	i, err := sc.Int64()
	return i, wrapNil(err)
}

//DeleteSubstream deletes the given substream from the stream
func (rc *RedisConnection) DeleteSubstream(stream, substream string) error {
	return wrapNil(rc.subdeleteScript.Run(rc.redis, []string{stream + ":" + substream + ":s", stream + ":m"},
		[]string{substream}).Err())
}

//DeleteStream removes an entire stream and all substreams from redis
func (rc *RedisConnection) DeleteStream(stream string) error {
	//First we must check all the substreams. to do this, we list the keys in metadata
	keys, err := rc.redis.HKeys(stream + ":m").Result()
	if err != nil {
		return err
	}

	for i := range keys {
		if len(keys[i]) > 7 && strings.HasPrefix(keys[i], "length:") {
			rc.DeleteSubstream(stream, keys[i][7:len(keys[i])])
		}
	}

	return rc.redis.Del(stream+":m", stream+"::s").Err()
}

//Clear the cache of all data - for testing purposes only, this obviously poofs all data in the cache, so
//no use in production environments please.
func (rc *RedisConnection) Clear() error {
	return rc.redis.FlushDb().Err()
}
