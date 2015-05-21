package timebatchdb

import (
	"errors"
	"math"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	//ErrorRedisWrongSize means that there might be corruption in the database. The functions detected a DatapointArray
	//returned from cache was of inconsistent size.
	ErrorRedisWrongSize = errors.New("Data array sized incorrectly")
	//ErrorRedisDNE is thrown when get is run on a key that does not exist in the database yet
	ErrorRedisDNE = errors.New("The key does not exist")
	//ErrorUnordered is thrown when the datapoints are not ordered by strictly increasing timestamp
	ErrorUnordered = errors.New("Datapoints not ordered by timestamp")
	//ErrorTimestamp is thrown when the data stream already contains at least one data ponit with a greater or equal timestamp to the ones being inserted
	ErrorTimestamp = errors.New("Greater or equal timestamp already exists for the stream")
)

//RedisCache is the redis-based cache of data which allows buffering of batches before committing to a long-term store
type RedisCache struct {
	cpool    *redis.Pool //The redis connection pool
	inserter *redis.Script
}

//Close all the connections to redis.
func (rc *RedisCache) Close() {
	rc.cpool.Close()
}

//Clear the cache of all data - for testing purposes only, this obviously disappears all data in the cache, so
//no use in production environments please.
func (rc *RedisCache) Clear() {
	c := rc.cpool.Get()
	c.Do("FLUSHDB")
	c.Close()
}

//EndIndex gets the index at which the key data currently ends. This is equivalent to querying the total number of datapoints
//for the given key in the entire timebatchdb
func (rc *RedisCache) EndIndex(key string) (endindex uint64, err error) {
	si, cl, err := rc.GetIndices(key)
	return si + cl, err
}

//StartIndex gets the index at which redis caching begins. This basically allows to check if all datapoints of the query can be
//satisfied by redis, or if querying the database is necessary
func (rc *RedisCache) StartIndex(key string) (startindex uint64, err error) {
	c := rc.cpool.Get()
	startindex, err = redis.Uint64(c.Do("GET", "{I}>"+key))
	c.Close()
	if err == redis.ErrNil { //If it returns nil, it means that the index simply DNE - so we just use 0
		err = nil
	}
	return startindex, err
}

//CacheLength returns the number of datapoints cached for the given key
func (rc *RedisCache) CacheLength(key string) (clen uint64, err error) {
	c := rc.cpool.Get()
	defer c.Close()
	return redis.Uint64(c.Do("LLEN", key))
}

//GetIndices returns the startindex and the number of datapoints currently cached in redis
func (rc *RedisCache) GetIndices(key string) (startindex uint64, cachelength uint64, err error) {
	c := rc.cpool.Get()
	c.Send("MULTI")
	c.Send("GET", "{I}>"+key)
	c.Send("LLEN", key)
	vals, err := redis.Values(c.Do("EXEC"))
	startindex, err = redis.Uint64(vals[0], err)
	if err == redis.ErrNil { //If it returns nil, it means that the index simply DNE - so we just use 0
		err = nil
	}
	cachelength, err = redis.Uint64(vals[1], err)
	return startindex, cachelength, err
}

//GetMostRecent gets the most recently inserted datapoint from the cache
func (rc *RedisCache) GetMostRecent(key string) (Datapoint, error) {
	c := rc.cpool.Get()
	v, err := redis.Values(c.Do("LRANGE", key, -1, -1))
	c.Close()
	if len(v) == 0 {
		return Datapoint{}, ErrorRedisDNE
	}
	dbytes, err := redis.Bytes(v[0], err)
	if err != nil {
		return Datapoint{}, err
	}
	dp, _ := DatapointFromBytes(dbytes)
	return dp, nil
}

//GetEndTime gets the time of the last inserted datapoint for the given key
func (rc *RedisCache) GetEndTime(key string) (t int64, err error) {
	c := rc.cpool.Get()
	t, err = redis.Int64(c.Do("GET", "{T}>"+key))
	c.Close()
	if err == redis.ErrNil { //If it returns nil, it means that the key DNE - so the timestamp is min
		return math.MinInt64, nil
	}
	return t, err
}

//GetOldest gets the oldest datapoint existing in the cache
func (rc *RedisCache) GetOldest(key string) (Datapoint, error) {
	c := rc.cpool.Get()
	v, err := redis.Values(c.Do("LRANGE", key, 0, 0))
	c.Close()
	if len(v) == 0 {
		return Datapoint{}, ErrorRedisDNE
	}
	dbytes, err := redis.Bytes(v[0], err)
	if err != nil {
		return Datapoint{}, err
	}
	dp, _ := DatapointFromBytes(dbytes)
	return dp, nil
}

//GetStartTime gets the timestamp of the oldest datapoint that exists in the cache
func (rc *RedisCache) GetStartTime(key string) (t int64, err error) {
	dp, err := rc.GetOldest(key)
	if err == ErrorRedisDNE {
		return math.MaxInt64, nil //We want to bound the starttime
	} else if err != nil {
		return 0, err
	}
	return dp.Timestamp(), err
}

//InsertSimple adds the DatapointArray to the end of the cache for the given key
func (rc *RedisCache) InsertSimple(key string, da *DatapointArray) (keysize int, err error) {
	c := rc.cpool.Get()
	//iterate the most recent timestamp
	c.Send("SET", "{T}>"+key, da.Datapoints[da.Len()-1].Timestamp())
	for i := 0; i < da.Len(); i++ {
		c.Send("RPUSH", key, da.Datapoints[i].Bytes())
	}
	keysize, err = redis.Int(c.Do("LLEN", key))
	c.Close()
	return keysize, err
}

//Insert checks timestamp orderedness, inserts the datpointArray, and performs a batch push, all in one Redis query.
func (rc *RedisCache) Insert(key string, da *DatapointArray, pushsize int) error {
	if !da.IsTimestampOrdered() || da.Len() == 0 {
		return ErrorUnordered
	}
	c := rc.cpool.Get()

	// go's variadic args are a bit annoying - we need the wrapper array for Do to accept the variadic array.
	args := []interface{}{key, "{T}>" + key, "{{READY_Q}}",
		da.Datapoints[0].Timestamp(), da.Datapoints[da.Len()-1].Timestamp(), pushsize}
	for i := 0; i < da.Len(); i++ {
		args = append(args, interface{}(da.Datapoints[i].Bytes()))
	}

	//Actually execute the inserter script (the code for it is located in the initializer)
	_, err := rc.inserter.Do(c, args...)

	c.Close()
	if err != nil && err.Error() == "TSM" {
		err = ErrorTimestamp
	}
	return err
}

//BatchPush adds the given key to the ready-queue - the queue of keys that have a batch ready to dump to disk storage
func (rc *RedisCache) BatchPush(key string) error {
	return rc.BatchPushMany(key, 1)
}

//BatchPushMany works same as BatchPush, but actually pushes a key the given number of times.
func (rc *RedisCache) BatchPushMany(key string, num int) error {
	c := rc.cpool.Get()
	for i := 0; i < num; i++ {
		c.Send("LPUSH", "{{READY_Q}}", key)
	}
	_, err := c.Do("")
	c.Close()
	return err
}

//BatchWait waits until there is a key in the ready-queue, and pops it
func (rc *RedisCache) BatchWait() (key string, err error) {
	c := rc.cpool.Get()
	keys, err := redis.Strings(c.Do("BRPOP", "{{READY_Q}}", 0)) //Blocking pop without timeout
	c.Close()
	if err != nil { //The array might not exist on error
		return "", err
	}
	return keys[1], nil
}

//BatchRemove marks the most recent n datapoints for the given key as "processed", and delete them from the cache.
//This means that the datapoints no longer need to be in the cache - they are already saved and committed
//to long term storage. This also increments the index marker in redis
func (rc *RedisCache) BatchRemove(key string, batchsize int) error {
	c := rc.cpool.Get()
	c.Send("MULTI")
	c.Send("LTRIM", key, batchsize, -1)
	c.Send("INCRBY", "{I}>"+key, batchsize) //Increment the "startindex" by the amount that was removed from the list
	_, err := c.Do("EXEC")
	return err
}

//BatchGet returns the last batchsize elements from the cache of the given key.
func (rc *RedisCache) BatchGet(key string, batchsize int) (da *DatapointArray, startindex uint64, err error) {
	c := rc.cpool.Get()
	defer c.Close()
	c.Send("GET", "{I}>"+key)
	c.Send("LRANGE", key, 0, batchsize-1)
	c.Flush()
	startindex, err = redis.Uint64(c.Receive())
	if err != nil && err != redis.ErrNil {
		return nil, startindex, err
	}
	values, err := redis.Values(c.Receive())
	if err != nil {
		return nil, startindex, err
	}
	//Create a DatapointArray from the response values
	dpa := make([]Datapoint, len(values))
	for i := 0; i < len(values); i++ {
		v, err := redis.Bytes(values[i], nil)
		if err != nil {
			return nil, startindex, err
		}
		dpa[i], _ = DatapointFromBytes(v)
	}
	da = NewDatapointArray(dpa)

	if batchsize > 0 && da.Len() < batchsize {
		return da, startindex, ErrorRedisWrongSize
	}
	return da, startindex, nil
}

//Get returns all of the elements in the cache for the given key
func (rc *RedisCache) Get(key string) (da *DatapointArray, startindex uint64, err error) {
	return rc.BatchGet(key, 0)
}

//GetByIndex returns the cache starting from the given index
func (rc *RedisCache) GetByIndex(key string, index uint64) (dr DataRange, startindex uint64, err error) {
	dp, si, err := rc.Get(key)
	if err != nil || si >= index {
		return dp, si, err
	} else if si+uint64(dp.Len()) <= index { //If index is out of bounds, return an empty range
		return EmptyRange{}, index, nil
	}
	return NewDatapointArray(dp.Datapoints[index-si:]), index, nil
}

//Delete all data associated with the given key stored within the cache.
func (rc *RedisCache) Delete(key string) error {
	c := rc.cpool.Get()
	defer c.Close()
	_, err := c.Do("DEL", key, "{I}>"+key, "{T}>"+key)
	return err
}

//DeletePrefix deletes all data associated with all keys which start with the given prefix.
//Warning: This is a pretty expensive operation, since redis has no built-in wildcard delete.
//Furthermore: this does not work with redis cluster.
func (rc *RedisCache) DeletePrefix(prefix string) (numberdeleted int, err error) {
	c := rc.cpool.Get()
	defer c.Close()
	//Redis does not have an explicit wildcard deletion method, so we eval a lua script
	//which lists the keys matching the prefix and deletes all of them.
	//This is not a very efficient method of doing things, since it requires iterating through all
	//of the keys in redis to match the wildcard. I would hope that we figure out something better some time
	//in the future.
	//The first keys call looks for timestamp, because if the array is empty, the key is not returned,
	//so the data is not deleted... It took me a good 3 hours to figure out why this wasn't deleting properly.
	return redis.Int(c.Do("EVAL", `local keys = redis.call('keys','{T}>' .. ARGV[1])
		local key = ""
		for i=1,#keys do
			key = string.sub(keys[i],5)
			redis.call('DEL', key, '{T}>' .. key, '{I}>' .. key)
		end
		return #keys`, 0, prefix+"*"))
}

//OpenRedisCache opens the redis cache given the URL to the server. The err parameter allows daisychains of errors
func OpenRedisCache(url string, err error) (*RedisCache, error) {
	if err != nil {
		return nil, err
	}
	rp := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", url)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	//Check if we can connect to redis
	c := rp.Get()
	err = c.Err()
	c.Close()
	if err != nil {
		rp.Close()
		return nil, err
	}

	//Load the inserter script. This script is used for fast insert.
	//This command first checks for timestamp validity. If valid, it sets the end time to the timestamp of
	//the most recent datapoint. Then, it pushes all the datapoints into the key's list in chunks, because
	//redis lua has an argument limit. Lastly, if the number of datapoints is large, perform a batch push of
	//the key, so that the database writer can notice the batch. There is one batch push for each batchsize.
	inserter := redis.NewScript(3, `local endtime = tonumber(redis.call('get',KEYS[2]))
	if (endtime~=nil and endtime >= tonumber(ARGV[1])) then
		return {["err"]="TSM"}
	end
	redis.call('set',KEYS[2],ARGV[2])
	for i=4,#ARGV,5000 do
		redis.call('rpush',KEYS[1],unpack(ARGV,i,math.min(i+4999,#ARGV)))
	end
	local datanum = tonumber(redis.call('llen',KEYS[1]))
	if datanum >= tonumber(ARGV[3]) then
		for i=1,math.floor((#ARGV-3)/tonumber(ARGV[3])) do
			redis.call('lpush',KEYS[3],KEYS[1])
		end
		if (#ARGV-3)%tonumber(ARGV[3])+(datanum-#ARGV+3)%tonumber(ARGV[3]) >= tonumber(ARGV[3]) then
			redis.call('lpush',KEYS[3],KEYS[1])
		end
	end`)

	return &RedisCache{rp, inserter}, nil
}
