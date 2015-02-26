package timebatchdb

import (
    "github.com/garyburd/redigo/redis"
    "time"
    )

type HotCache interface {
    InsertOne(key string, dp Datapoint) error  //Inserts
    Len(key string) int
    Get(key string) DataRange
    GetIndex(key string, startindex uint64) DataRange
    Close()
}

type RedisRange struct {
    pool *redis.Pool //The redis connection pool
    key string          //The key to query
    startindex uint64   //The index to start range on
    da *DatapointArray  //The datapoint array (queried on init)
}

func (r *RedisRange) Close() {

}

func (r *RedisRange) Next() *Datapoint {
    if r.da == nil {
        return nil
    }
    return r.da.Next()
}

func (r *RedisRange) Init() {
    c := r.pool.Get()

    c.Send("MULTI")
    c.Send("GET","i>"+r.key)
    c.Send("LRANGE",r.key,0,-1)

    values, err := redis.Values(c.Do("EXEC"))
    c.Close()

    if err!=nil || len(values)!=2 {
        return  //We failed
    }

    //Get the data index - when key DNE returns 0, which is fine by me
    dataindex,_ := redis.Uint64(values[0],nil)

    starti := 0
    if (dataindex < r.startindex) {
        starti += int(r.startindex - dataindex)
    }

    values,err = redis.Values(values[1],nil)

    if err!= nil || len(values)==0 {
        return  //We failed
    }

    if (len(values)<= starti) {
        return //No values fit in the given range
    }

    //Create a DatapointArray from the response values
    dpa := make([]Datapoint,len(values)-starti)
    for i := 0 ; i+starti < len(values); i++ {
        v,err := redis.Bytes(values[i+starti],nil)
        if err != nil {
            return
        }
        dpa[i],_ = DatapointFromBytes(v)
    }
    r.da =  NewDatapointArray(dpa)
}



type RedisCache struct {
    cpool *redis.Pool //The redis connection pool
}

func (hc *RedisCache) Close() {
    hc.cpool.Close()
}

//Clears the cache of all data
func (hc *RedisCache) Clear() {
    c := hc.cpool.Get()
    c.Do("FLUSHDB")
    c.Close()
}

func (hc *RedisCache) InsertOne(key string, dp Datapoint) error {
    c := hc.cpool.Get()
    _,err := c.Do("RPUSH",key,dp.Bytes())
    c.Close()
    return err
}

func (hc *RedisCache) Len(key string) int {
    c := hc.cpool.Get()
    r,err := redis.Int(c.Do("LLEN",key))
    c.Close()
    if err!= nil {
        return -1
    }
    return r
}

func (hc *RedisCache) Get(key string) DataRange {
    return hc.GetIndex(key,0)
}

func (hc *RedisCache) GetIndex(key string, index uint64) DataRange {
    return &RedisRange{hc.cpool,key,index,nil}
}

func OpenRedisCache(url string) (*RedisCache,error) {
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
