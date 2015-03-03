package timebatchdb

import (
    "github.com/garyburd/redigo/redis"
    "time"
    "errors"
    )

var (
    ERROR_REDIS_WRONGSIZE = errors.New("Data array sized incorrectly")
)


//The redis-based cache of data which allows buffering of batches before committing to timebatchdb
type RedisCache struct {
    cpool *redis.Pool  //The redis connection pool
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

//Get the index at which the key data currently ends. This is equivalent to querying the total number of datapoints
//for the given key in the entire timebatchdb
func (rc *RedisCache) EndIndex(key string) (endindex uint64, err error) {
    si,cl,err := rc.GetIndices(key)
    return si+cl,err
}

//Get the index at which redis caching begins. This basically allows to check if all datapoints of the query can be
//satisfied by redis, or if querying the database is necessary
func (rc *RedisCache) StartIndex(key string) (startindex uint64, err error) {
    c := rc.cpool.Get()
    startindex,err = redis.Uint64(c.Do("GET","{I}>"+key))
    c.Close()
    if err==redis.ErrNil {  //If it returns nil, it means that the index simply DNE - so we just use 0
        err=nil
    }
    return startindex,err
}

//Returns the number of datapoints cached for the given key
func (rc *RedisCache) CacheLength(key string) (clen uint64, err error) {
    c := rc.cpool.Get()
    defer c.Close()
    return redis.Uint64(c.Do("LLEN",key))
}

//Returns the startindex and the number of datapoints currently cached in redis
func (rc *RedisCache) GetIndices(key string) (startindex uint64,cachelength uint64, err error) {
    c := rc.cpool.Get()
    c.Send("MULTI")
    c.Send("GET","{I}>"+key)
    c.Send("LLEN",key)
    vals,err := redis.Values(c.Do("EXEC"))
    startindex,err = redis.Uint64(vals[0],err)
    if err==redis.ErrNil {  //If it returns nil, it means that the index simply DNE - so we just use 0
        err=nil
    }
    cachelength,err = redis.Uint64(vals[1],err)
    return startindex,cachelength,err
}



//Insert the DatapointArray to the end of the cache for the given key
func (rc *RedisCache) Insert(key string,da *DatapointArray) (keysize int, err error) {
    c := rc.cpool.Get()
    for i:= 0 ; i < da.Len() ; i++ {
        c.Send("RPUSH",key,da.Datapoints[i].Bytes())
    }
    keysize,err = redis.Int(c.Do("LLEN",key))
    c.Close()
    return keysize,err
}

//Adds the given key to the ready-queue - the queue of keys that have a batch ready to dump to disk storage
func (rc *RedisCache) BatchPush(key string) error {
    c := rc.cpool.Get()
    _,err := c.Do("LPUSH","{{READY_Q}}",key)
    c.Close()
    return err
}

//Waits until there is a key in the ready-queue, and pops it
func (rc *RedisCache) BatchWait() (key string, err error) {
    c := rc.cpool.Get()
    keys,err := redis.Strings(c.Do("BRPOP","{{READY_Q}}",0))    //Blocking pop without timeout
    c.Close()
    return keys[1],err
}

//Mark the most recent n datapoints for the given key as "processed", and delete them from the cache.
//This means that the datapoints no longer need to be in the cache - they are already saved and committed
//to long term storage. This also increments the index marker in redis
func (rc *RedisCache) BatchRemove(key string, batchsize int) error {
    c := rc.cpool.Get()
    c.Send("MULTI")
    c.Send("LTRIM",key,batchsize,-1)
    c.Send("INCRBY","{I}>"+key,batchsize) //Increment the "startindex" by the amount that was removed from the list
    _,err := c.Do("EXEC")
    return err
}

//Returns the last batchsize elements from the cache of the given key.
func (rc *RedisCache) BatchGet(key string, batchsize int) (da *DatapointArray, startindex uint64, err error) {
    c := rc.cpool.Get()
    defer c.Close()
    c.Send("GET","{I}>"+key)
    c.Send("LRANGE",key,0,batchsize-1)
    c.Flush()
    startindex,err = redis.Uint64(c.Receive())
    if err!=nil && err!=redis.ErrNil {
        return nil,startindex,err
    }
    values,err := redis.Values(c.Receive())
    if err!=nil {
        return nil,startindex,err
    }
    //Create a DatapointArray from the response values
    dpa := make([]Datapoint,len(values))
    for i := 0 ; i < len(values); i++ {
        v,err := redis.Bytes(values[i],nil)
        if err != nil {
            return nil,startindex,err
        }
        dpa[i],_ = DatapointFromBytes(v)
    }
    da =  NewDatapointArray(dpa)

    if (batchsize>0 && da.Len() < batchsize) {
        return da,startindex,ERROR_REDIS_WRONGSIZE
    }
    return da,startindex,nil
}


//Returns all of the elements in the cache for the given key
func (rc *RedisCache) Get(key string) (da *DatapointArray, startindex uint64, err error) {
    return rc.BatchGet(key,0)
}

//Opens the redis cache given the URL to the server
func OpenRedisCache(url string) (*RedisCache, error) {
    rp := &redis.Pool{
        MaxIdle: 3,
        IdleTimeout: 240 * time.Second,
        Dial: func () (redis.Conn, error) {
            c,err := redis.Dial("tcp",url)
            if err != nil {
                return nil, err
            }
            return c,err
        },
        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            _, err := c.Do("PING")
            return err
        },
        }

    //Check if we can connect to redis
    c := rp.Get()
    err := c.Err()
    c.Close()
    if err!=nil {
        rp.Close()
        return nil,err
    }

    return &RedisCache{rp},nil
}
