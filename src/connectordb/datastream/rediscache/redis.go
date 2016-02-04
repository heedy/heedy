/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package rediscache

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gopkg.in/redis.v3"

	"connectordb/datastream"
)

/*
The structure in Redis is as follows:

Each device has an hset. Each stream has a list. Datapoints are inserted one per list element
all marshalled into bytes using messagepack.

Suppose that we have a stream "mystream" belonging to "mydevice".

	'{mydevice}:mystream:' = [(datapoint),(datapoint),(datapoint)]

	'{mydevice}' = {
		'endtime:mystream:' : the most recent timestamp of inserted data
		'starttime:mystream:' : the first timestamp of data in redis
		'length:mystream:' : the total number of datapoints in the stream (overall)
	}


Notice that most elements end in : (or have extra :).
This is because streams can have substreams. A stream 'mystream' with a downlink substream 'd'
would look like this:

	'{mydevice}:mystream:' = [(datapoint),(datapoint),(datapoint)]
	'{mydevice}:mystream:d' = [(datapoint),(datapoint),(datapoint)]

	'{mydevice}' = {
		'endtime:mystream:' : the most recent timestamp of inserted data
		'starttime:mystream:' : the first timestamp of data in redis
		'length:mystream:' : the total number of datapoints in the stream (overall)
		'size:mystream': the size of the stream in bytes

		'endtime:mystream:d' : the most recent timestamp of inserted data in 'd'
		'starttime:mystream:d' : the first timestamp of data in redis for substream
		'length:mystream:d' : the total number of datapoints in the substream 'd'
		'size:mystream:d': the size of the substream in bytes

		'size': the total size of all data that mydevice holds in bytes
	}

Lastly, notice the {} around mydevice - this is for redis cluster hashing.
It allows all keys relevant to a device to be on the same redis instance,
which is exploited heavily in the scripts.

Lastly, there is the "batch-list" which is a lsit of batches that are waiting to be written to
the database. These batches are handled by the batch-writer processes, which is a simple writer
in a single instance, and a more complicated machinery when working on clusters.

*/

const (
	//RedisNilString is the string error that is returned if redis gives nil
	RedisNilString = "redis: nil"

	//The insert script does the following:
	//It is given 3 keys:
	//1	stream key - the key where a list of chunks has been inserted
	//2	metadata key - the key where the stream's metadata is stored
	//3	batch writer key - the key to which to write batches. If == stream key, doesn't write batches
	//Of the arguments, it is given:
	//1	subpath - the name of the stream in "stream:substream" format
	//2	starttime - the start time of the datapoints
	//3	endtime - the end time of the datapoints
	//4	restamp - whether to restamp datapoints if inconsistent timestamps
	//5	batchsize - the number of datapoints which constitute a batch
	//6	datasize - the size of the currently inserted array in bytes
	//7	maxdevicesize - the maximum number of bytes to permit in a device. =0 means unlimited
	//8	maxstreamsize - the maximum number of bytes to permit in a stream. =0 means unlimited
	//	... array of the datapoints to be inserted ...
	insertScript = `
		-- Check to make sure we don't go over the size limits for device and stream
		if (ARGV[7] ~= '0') then
			local device_size = tonumber(redis.call('hget',KEYS[2], 'size')) or 0
			if (device_size + tonumber(ARGV[6]) > tonumber(ARGV[7])) then
				return {["err"]="Insert Failed: Exceeded device size limit"}
			end
		end
		if (ARGV[8] ~= '0') then
			local stream_size = tonumber(redis.call('hget',KEYS[2], 'size:' .. ARGV[1])) or 0
			if (stream_size + tonumber(ARGV[6]) > tonumber(ARGV[8])) then
				return {["err"]="Insert Failed: Exceeded stream size limit"}
			end
		end

		-- Make sure that the timestamps are increasing
		local stream_endtime = tonumber(redis.call('hget',KEYS[2], 'endtime:' .. ARGV[1])) or 0
		if (stream_endtime > tonumber(ARGV[2])) then
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

			for i=#ARGV,9,-1 do
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
		redis.call('hincrby',KEYS[2], 'length:' .. ARGV[1], #ARGV - 8)
		-- Set the stream size
		redis.call('hincrby',KEYS[2], 'size:' .. ARGV[1], ARGV[6])
		-- Set the device total size
		redis.call('hincrby',KEYS[2], 'size', ARGV[6])

		-- Insert the datapoints into the stream - redis lua has some weird stuff about the maximum
		-- number of arguments to a function - we avoid this by manually splitting insert into chunks
		for i=9,#ARGV,5000 do
			redis.call('rpush',KEYS[1], unpack(ARGV,i,math.min(i+4999,#ARGV)))
		end

		-- Check to see if we should write a batch
		local streamlength = tonumber(redis.call('hget',KEYS[2], 'length:' .. ARGV[1]))
		local batchindex = tonumber(redis.call('hget',KEYS[2], 'batchindex:' .. ARGV[1])) or 0
		local batchsize = tonumber(ARGV[5])
		if (streamlength > batchindex + batchsize) then
			-- Yep, write at least one batch
			local batchnum = math.floor((streamlength-batchindex)/batchsize)
			local batches = {}
			for i=batchindex,streamlength-batchsize,batchsize do
				table.insert(batches,KEYS[1] .. ":" .. i .. ":" .. (i+batchsize))
			end
			redis.call('lpush',KEYS[3],unpack(batches))
			redis.call('hset',KEYS[2], 'batchindex:' .. ARGV[1], batchindex+batchsize*batchnum)
		end

		return streamlength
	`

	//The subdelete script deletes a given substream.
	//Given 2 keys:
	//	the stream key
	//	metadata key
	//In arguments it is given:
	//	the substream to delete
	subdeleteScript = `
		-- Delete the stream list
		redis.call('del',KEYS[1])

		-- Subtract the stream size from the full device size
		local stream_size = tonumber(redis.call('hget',KEYS[2], 'size:' .. ARGV[1])) or 0
		redis.call('hincrby', KEYS[2], 'size', -stream_size)

		-- Remove metadata
		redis.call('hdel',KEYS[2],'endtime:' .. ARGV[1], 'length:' .. ARGV[1], 'starttime:' .. ARGV[1], 'batchindex:' .. ARGV[1], 'size:' .. ARGV[1])
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
	//	The stream metadata subpath
	//	i1 startindex
	//	i2 endindex
	rangeScript = `
		local redislength = tonumber(redis.call('llen',KEYS[1]))
		local streamlength = tonumber(redis.call('hget',KEYS[2], 'length:' .. ARGV[1]))
		local i1 = tonumber(ARGV[2])
		local i2 = tonumber(ARGV[3])

		if (redislength==nil or streamlength==nil) then
			-- The stream doesn't exist. return 0,0 if a end-relative range
			if (i1<=0) then
				return {0,0}
			end
			-- otherwise, it is an invalid range
			return {["err"]="Invalid index range."}
		end

		-- If the indices are from the end, set their values
		if (i1<0) then
			i1 = streamlength + i1
			-- Negative indices further than bound should read from beginning of stream instead
			if (i1<0) then
				i1 = 0
			end
		end
		if (i2<=0) then
			i2 = streamlength + i2
		end

		-- If the second index is out of bounds, just return what we have
		if (i2 > streamlength) then
			i2 = streamlength
		end

		if (i2 < i1) then
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
	//	The stream path
	//	index to trim to (ie, keep datapoints AFTER index)
	trimScript = `
		local startindex = tonumber(redis.call('hget',KEYS[2], 'length:' .. ARGV[1])) - tonumber(redis.call('llen',KEYS[1]))
		local i = tonumber(ARGV[2])

		if (i > startindex) then
			-- We can trim data from the end
			redis.call('ltrim',KEYS[1], i - startindex, -1)
			return 'ok'
		end
	`
)

var (
	//ErrTimestamp is returned when trying to insert old timestamps
	ErrTimestamp = errors.New("Greater timestamp already exists for the stream. Insert Failed.")

	//ErrWTF is returned when an internal assertion fails - it should not happen. Ever.
	ErrWTF = errors.New("Something is seriously wrong. A internal assertion failed.")
)

//Quite annoyingly, go-redis does not give an interface using which we can connect to redis. We therefore manually create one
type redisConnection interface {
	LRange(key string, start, stop int64) *redis.StringSliceCmd
	HGet(key, field string) *redis.StringCmd
	HKeys(key string) *redis.StringSliceCmd
	Del(keys ...string) *redis.IntCmd
	FlushDb() *redis.StatusCmd

	//This enables us to use a redis.Script object
	Eval(script string, keys []string, args []string) *redis.Cmd
	EvalSha(sha1 string, keys []string, args []string) *redis.Cmd
	ScriptExists(scripts ...string) *redis.BoolSliceCmd
	ScriptLoad(script string) *redis.StringCmd

	BRPopLPush(source, destination string, timeout time.Duration) *redis.StringCmd

	Close() error
}

//RedisConnection is the connection to redis server
type RedisConnection struct {
	Redis     redisConnection
	BatchSize int64

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

//Converts the stream to an int
func stringID(id int64) string {
	return strconv.FormatInt(id, 32)
}

func streamKey(hash, stream, substream string) string {
	return "{" + hash + "}" + stream + ":" + substream
}

//Many of the redis scripts use the same exact keys - this just abstracts that away
func scriptkeys(hash, stream, substream string) []string {
	return []string{streamKey(hash, stream, substream), "{" + hash + "}"}
}

//Close the redis cluster connection
func (rc *RedisConnection) Close() error {
	return rc.Redis.Close()
}

//NewRedisConnection creates a new connection to a redis server. This is a single-node
//	redis connection
func NewRedisConnection(opt *redis.Options) (rc *RedisConnection, err error) {
	rclient := redis.NewClient(opt)

	_, err = rclient.Ping().Result()

	return &RedisConnection{
		Redis:           rclient,
		BatchSize:       250,
		insertScript:    redis.NewScript(insertScript),
		subdeleteScript: redis.NewScript(subdeleteScript),
		rangeScript:     redis.NewScript(rangeScript),
		trimScript:      redis.NewScript(trimScript),
	}, err
}

//Get returns all the datapoints cached associated with the given stream/substream
func (rc *RedisConnection) Get(hash, stream, substream string) (dpa datastream.DatapointArray, err error) {
	sa, err := rc.Redis.LRange(streamKey(hash, stream, substream), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return datastream.DatapointArrayFromDataStrings(sa)

}

//Insert datapoint array, writing batches to batchkey
func (rc *RedisConnection) Insert(batchkey, hash, stream, substream string, dpa datastream.DatapointArray, restamp bool, maxDeviceSize, maxStreamSize int64) (streamlength int64, err error) {
	//remember the number of args here
	args := make([]string, 8+len(dpa))

	args[0] = stream + ":" + substream
	args[1] = strconv.FormatFloat(dpa[0].Timestamp, 'G', -1, 64)
	args[2] = strconv.FormatFloat(dpa[len(dpa)-1].Timestamp, 'G', -1, 64)
	if restamp {
		args[3] = "1"
	} else {
		args[3] = "0"
	}
	args[4] = strconv.FormatInt(rc.BatchSize, 10)
	// Arg 5 will be inserted after finding data size
	args[6] = strconv.FormatInt(maxDeviceSize, 10)
	args[7] = strconv.FormatInt(maxStreamSize, 10)

	datasize := int64(0)
	for i := range dpa {
		b, err := dpa[i].Bytes()
		if err != nil {
			return 0, err
		}
		datasize += int64(len(b))
		args[i+8] = string(b)
	}

	args[5] = strconv.FormatInt(datasize, 10)

	r, err := rc.insertScript.Run(rc.Redis, []string{streamKey(hash, stream, substream), "{" + hash + "}", batchkey}, args).Result()

	if err != nil {
		return 0, err
	}

	return r.(int64), err
}

//StreamLength returns the stream's length
func (rc *RedisConnection) StreamLength(hash, stream, substream string) (int64, error) {
	sc := rc.Redis.HGet("{"+hash+"}", "length:"+stream+":"+substream)

	i, err := sc.Int64()
	return i, wrapNil(err)
}

// StreamSize returns the stream's size in bytes
func (rc *RedisConnection) StreamSize(hash, stream, substream string) (int64, error) {
	sc := rc.Redis.HGet("{"+hash+"}", "size:"+stream+":"+substream)

	i, err := sc.Int64()
	return i, wrapNil(err)
}

//DeleteSubstream deletes the given substream from the stream
func (rc *RedisConnection) DeleteSubstream(hash, stream, substream string) error {
	return wrapNil(rc.subdeleteScript.Run(rc.Redis, scriptkeys(hash, stream, substream),
		[]string{stream + ":" + substream}).Err())
}

//DeleteStream removes an entire stream and all substreams from redis
//WARNING: This is not atomic. If a substream is created in the middle of deletion,
//the substream's data will become corrupted and not cleaned up. I am assuming that
//this will become disallowed somehow through the higher level interface
func (rc *RedisConnection) DeleteStream(hash string, stream string) (err error) {
	//First we must check all the substreams. to do this, we list the keys in hash
	keys, err := rc.Redis.HKeys("{" + hash + "}").Result()
	if err != nil {
		return err
	}

	for i := range keys {
		if len(keys[i]) > 7 && strings.HasPrefix(keys[i], "length:"+stream+":") {
			err := rc.DeleteSubstream(hash, stream, keys[i][8+len(stream):])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//DeleteHash removes all streams within a hash
func (rc *RedisConnection) DeleteHash(hash string) (err error) {
	//First we must check all the substreams. to do this, we list the keys in hash
	keys, err := rc.Redis.HKeys("{" + hash + "}").Result()
	if err != nil {
		return err
	}

	todelete := []string{"{" + hash + "}"}

	for i := range keys {
		if len(keys[i]) > 7 && strings.HasPrefix(keys[i], "length:") {
			todelete = append(todelete, "{"+hash+"}"+keys[i][7:])
		}
	}

	return rc.Redis.Del(todelete...).Err()
}

// HashSize returns the device size in bytes
func (rc *RedisConnection) HashSize(hash string) (int64, error) {
	sc := rc.Redis.HGet("{"+hash+"}", "size")

	i, err := sc.Int64()
	return i, wrapNil(err)
}

//TrimStream clears all datapoints up to the index from redis, after they are written
//to long term storage, so that they don't take up space.
func (rc *RedisConnection) TrimStream(hash, stream, substream string, index int64) error {
	return wrapNil(rc.trimScript.Run(rc.Redis, scriptkeys(hash, stream, substream), []string{stream + ":" + substream, strconv.FormatInt(index, 10)}).Err())
}

//NextBatch waits for the next batch, and pushes it into the "in progress queue"
func (rc *RedisConnection) NextBatch(batchlist, progresslist string) (string, error) {
	return rc.Redis.BRPopLPush(batchlist, progresslist, 0).Result()
}

//GetList gets all the elements of the given list
func (rc *RedisConnection) GetList(listkey string) ([]string, error) {
	return rc.Redis.LRange(listkey, 0, -1).Result()
}

//DeleteKey removes the given key from the database
func (rc *RedisConnection) DeleteKey(key string) error {
	return rc.Redis.Del(key).Err()
}

//ReadBatch reads a batch of data given the batch string (as returned from NextBatch)
func (rc *RedisConnection) ReadBatch(batchstring string) (b *datastream.Batch, err error) {
	stringarray := strings.Split(batchstring, ":")

	if len(stringarray) != 4 {
		return nil, ErrWTF
	}
	substream := stringarray[1]
	i1, err := strconv.ParseInt(stringarray[2], 10, 64)
	if err != nil {
		return nil, err
	}
	i2, err := strconv.ParseInt(stringarray[3], 10, 64)
	if err != nil {
		return nil, err
	}

	stringarray = strings.Split(stringarray[0], "}")
	if len(stringarray) != 2 {
		return nil, ErrWTF
	}

	stream := stringarray[1]
	hash := stringarray[0][1:]

	dpa, i1, i2, err := rc.Range(hash, stream, substream, i1, i2)
	if err != nil {
		return nil, err
	}

	return &datastream.Batch{
		Device:     hash,
		Stream:     stream,
		Substream:  substream,
		StartIndex: i1,
		Data:       dpa,
	}, nil
}

//Range either gets the entire given range of data from redis, or notifies the indices of data to use in terms
//of the entire stream. The indices can be negative. For example, i1 negative means "from the end" - i2 = 0 means "to the end",
//so a range of -1,0 returns the most recent datapoint, -3,-1 returns 2 of the 3 most recent datapoints, 5,-1 returns index 5 to the
//second to last, and so forth. It is python-like indexing.
func (rc *RedisConnection) Range(hash, stream, substream string, index1, index2 int64) (dpa datastream.DatapointArray, i1, i2 int64, err error) {
	res, err := rc.rangeScript.Run(rc.Redis, scriptkeys(hash, stream, substream), []string{stream + ":" + substream, strconv.FormatInt(index1, 10), strconv.FormatInt(index2, 10)}).Result()
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
	dpa, err = datastream.DatapointArrayFromDataStrings(stringresult)
	return dpa, i1, i2, err

}

//Clear the cache of all data - for testing purposes only, this obviously poofs all data in the cache, so
//no use in production environments please. WARNING: Undefined behavior for redis cluster
func (rc *RedisConnection) Clear() error {
	return rc.Redis.FlushDb().Err()
}
