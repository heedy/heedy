package streamdb

import (
	"time"

	"github.com/hashicorp/golang-lru"
)

type cacheElement struct {
	loadtime int64 //The time when the element was loaded from database
	element  interface{}
}

//TimedCache is an lru cache with an additional expiration time for the data in each element
//It is not a very clever cache, and does not factor in time expiration in lru
type TimedCache struct {
	c          *lru.Cache
	expireTime int64
}

//NewTimedCache returns the TimedCache with given parameters
func NewTimedCache(size int, expireTime int64, err error) (TimedCache, error) {
	if err != nil {
		return TimedCache{}, err
	}
	l, err := lru.New(size)
	return TimedCache{l, expireTime}, err
}

//Add a new element to the cache
func (tc *TimedCache) Add(key string, value interface{}) {
	tc.c.Add(key, cacheElement{time.Now().Unix(), value})
}

//Get an element if it is cached and not yet expired
func (tc *TimedCache) Get(key string) (value interface{}, ok bool) {
	v, ok := tc.c.Get(key)
	if !ok {
		return nil, false
	}
	ce := v.(cacheElement)
	if ce.loadtime+tc.expireTime <= time.Now().Unix() {
		//The element is expired - it shouldn't be in the cache
		tc.c.Remove(key)
		return nil, false
	}
	return ce.element, true
}

//Removes the given key from the cache
func (tc *TimedCache) Remove(key string) {
	tc.c.Remove(key)
}

//Purge clears all elements in the cache
func (tc *TimedCache) Purge() {
	tc.c.Purge()
}
