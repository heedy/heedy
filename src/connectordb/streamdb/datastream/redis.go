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

	'{mystream}::s' = [(datapoint),(datapoint),(datapoint)]

	'{mystream}:m' = {
		'endtime:' : the most recent timestamp of inserted data
		'starttime:' : the first timestamp of data in redis
		'length:' : the total number of datapoints in the stream (overall)
	}

Notice that most elements end in : (or have extra :).
This is because streams can have substreams. A stream 'mystream' with a downlink substream 'd'
would look like this:

	'{mystream}::s' = [(datapoint),(datapoint),(datapoint)]
	'{mystream}:d:s' = [(datapoint),(datapoint),(datapoint)]

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
			-- it encodes 5.0 as int - meaning that floats can lose floatiness once repacked
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
		-- number of arguments to a function - we avoid this by manually splitting insert into chunks
		for i=5,#ARGV,5000 do
			redis.call('rpush',KEYS[1], unpack(ARGV,i,math.min(i+4999,#ARGV)))
		end

		return redis.call('hget',KEYS[2], 'length:' .. ARGV[1])
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

	//The range script returns the data from the given range of datapoints, and the 2 indices,
	//if the data is not in redis, returns just the index range.
	//While this could be done client side rather than redis-side, I wanted getting the most recent
	//datapoints to be as fast an operation as possible - which means that saving a round-trip if possible
	//is something desirable.
	//TODO(daniel): One thing that still needs to be benchmarked is how having this script affects redis itself,
	//	we are moving some processing to a single-threaded redis server, which might not be a good idea in a
	//	multinode setup
	//Given 2 keys:
	//	the stream key
	//	the metadata key
	//In arguments it is given:
	//	The substream to read
	//	i1 startindex
	//	i2 endindex
	rangeScript = `
		local redislength = tonumber(redis.call('llen',KEYS[1]))
		local streamlength = tonumber(redis.call('hget',KEYS[2], 'length:' .. ARGV[1]))
		local i1 = tonumber(ARGV[2])
		local i2 = tonumber(ARGV[3])

		if (redislength==nil or streamlength==nil) then
			-- The stream doesn't exist. If i1=0 and i2 > 0, then return 0,0
			if (i1==0 and i2 > 0) then
				return {0,0}
			end
			-- otherwise, it is an invalid range
			return {["err"]="Invalid index range."}
		end

		-- If the indices are from the end, set their values
		if (i1<0) then
			i1 = streamlength + i1
		end
		if (i2<=0) then
			i2 = streamlength + i2
		end

		-- If the second index is out of bounds, just return what we have
		if (i2 > streamlength) then
			i2 = streamlength
		end

		if (i2<=i1 or i1 < 0) then
			return {["err"]="Invalid index range."}
		end

		-- Now check if we can service this request
		local startloc = streamlength - redislength
		if (i1 < startloc) then
			-- We can't, so return the two indices only
			return {i1,i2}
		end

		-- Return the datapoints, and the two indices at the end
		local result = redis.call('lrange',KEYS[1], i1 - startloc, i2 - startloc-1)
		table.insert(result,i1)
		table.insert(result,i2)
		return result
	`

	//The script trims datapoints in a stream to start at the given index
	//The reason this is a script rather than doing it client-side is that
	//doing multiple queries allows another process to trim the array in-between,
	//which could lead to dataloss.
	//Given 2 keys:
	//	the stream key
	//	the metadata key
	//In arguments it is given:
	//	The substream to read
	//	index to trim to (ie, keep datapoints AFTER index)
	trimScript = `
		local startindex = tonumber(redis.call('hget',KEYS[2], 'length:' .. ARGV[1])) - tonumber(redis.call('llen',KEYS[1]))
		local i = tonumber(ARGV[2])

		if (i > startindex) then
			-- We can trim data from the end
			redis.call('ltrim',KEYS[1], i - startindex, -1)
		end
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
	rangeScript     *redis.Script
	trimScript      *redis.Script
}

//If redis returns nil, that is handled as an error in the redis library - this allows to wrap commands
//that we know return nil
func wrapNil(err error) error {
	if err != nil && err.Error() == RedisNilString {
		return nil
	}
	return err
}

//Many of the redis scripts use the same exact keys - this just abstracts that away
func scriptkeys(stream, substream string) []string {
	return []string{stream + ":" + substream + ":s", stream + ":m"}
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
		rangeScript:     redis.NewScript(rangeScript),
		trimScript:      redis.NewScript(trimScript),
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
func (rc *RedisConnection) Insert(stream, substream string, dpa DatapointArray, restamp bool) (streamlength int64, err error) {
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
			return 0, err
		}
		args[i+4] = string(b)
	}

	r, err := rc.insertScript.Run(rc.redis, scriptkeys(stream, substream), args).Result()

	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(r.(string), 10, 64)
}

//StreamLength returns the stream's length
func (rc *RedisConnection) StreamLength(stream, substream string) (int64, error) {
	sc := rc.redis.HGet(stream+":m", "length:"+substream)

	i, err := sc.Int64()
	return i, wrapNil(err)
}

//DeleteSubstream deletes the given substream from the stream
func (rc *RedisConnection) DeleteSubstream(stream, substream string) error {
	return wrapNil(rc.subdeleteScript.Run(rc.redis, scriptkeys(stream, substream),
		[]string{substream}).Err())
}

//DeleteStream removes an entire stream and all substreams from redis
//WARNING: This is not atomic. If a substream is created in the middle of deletion,
//the substream's data will become corrupted and not cleaned up. I am assuming that
//this will become disallowed somehow through the higher level interface
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

//TrimStream clears all datapoints up to the index from redis, after they are written
//to long term storage, so that they don't take up space.
func (rc *RedisConnection) TrimStream(stream, substream string, index int64) error {
	return wrapNil(rc.trimScript.Run(rc.redis, scriptkeys(stream, substream), []string{substream, strconv.FormatInt(index, 10)}).Err())
}

//Range either gets the entire given range of data from redis, or notifies the indices of data to use in terms
//of the entire stream. The indices can be negative. For example, i1 negative means "from the end" - i2 = 0 means "to the end",
//so a range of -1,0 returns the most recent datapoint, -3,-1 returns 2 of the 3 most recent datapoints, 5,-1 returns index 5 to the
//second to last, and so forth. It is python-like indexing.
func (rc *RedisConnection) Range(stream, substream string, index1 int64, index2 int64) (dpa DatapointArray, i1, i2 int64, err error) {
	res, err := rc.rangeScript.Run(rc.redis, scriptkeys(stream, substream), []string{substream, strconv.FormatInt(index1, 10), strconv.FormatInt(index2, 10)}).Result()
	if err != nil {
		return nil, 0, 0, err
	}

	//The result is actually a string array
	result, ok := res.([]interface{})
	if !ok || len(result) < 2 {
		return nil, 0, 0, ErrWTF
	}

	//The last two are the indices
	i1 = result[len(result)-2].(int64)
	i2 = result[len(result)-1].(int64)

	//If that's it, that means the index was out of range in redis
	if len(result) == 2 {
		return nil, i1, i2, nil
	}

	//If not, that means that redis returned the range!
	stringresult := make([]string, len(result)-2)
	for i := 0; i < len(result)-2; i++ {
		stringresult[i] = result[i].(string)
	}
	dpa, err = DatapointArrayFromDataStrings(stringresult)
	return dpa, i1, i2, err

}

//Clear the cache of all data - for testing purposes only, this obviously poofs all data in the cache, so
//no use in production environments please.
func (rc *RedisConnection) Clear() error {
	return rc.redis.FlushDb().Err()
}
