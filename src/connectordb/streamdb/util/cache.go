package util

import (
	"strings"
	"sync"
	"time"

	"github.com/dkumor/golang-lru"
)

type cacheElement struct {
	loadTime int64  //The time when the element was loaded from database
	keyName  string //The name of the key (if known)
	element  interface{}
}

//TimedCache is an lru cache with an additional expiration time for the data in each element
//It is not a very clever cache, and does not factor in time expiration in lru
type TimedCache struct {
	sync.Mutex //Necessary to use the nameMap
	c          *lru.Cache
	expireTime int64
	nameMap    map[string]int64 //A map that goes from name to ID
}

//NewTimedCache returns the TimedCache with given parameters
func NewTimedCache(size int, expireTime int64, err error) (*TimedCache, error) {
	if err != nil {
		return nil, err
	}

	tc := &TimedCache{sync.Mutex{}, nil, expireTime, make(map[string]int64)}

	//Add an eviction
	tc.c, err = lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		tc.Lock()
		delete(tc.nameMap, value.(cacheElement).keyName)
		tc.Unlock()
	})
	return tc, err
}

//Set adds a new element to the cache
func (tc *TimedCache) Set(name string, id int64, value interface{}) {
	if _, keyName, ok := tc.GetByID(id); ok && keyName != "" && keyName != name {
		tc.UnlinkName(keyName)
	}
	tc.c.Add(id, cacheElement{time.Now().Unix(), name, value})
	tc.Lock()
	tc.nameMap[name] = id
	tc.Unlock()
}

//Update an element without changing the associated name
func (tc *TimedCache) Update(id int64, value interface{}) {
	_, keyName, _ := tc.GetByID(id)
	tc.c.Add(id, cacheElement{time.Now().Unix(), keyName, value})
}

//SetID sets a new element without knowing its name
func (tc *TimedCache) SetID(id int64, value interface{}) {
	if _, keyName, ok := tc.GetByID(id); ok && keyName != "" {
		tc.UnlinkName(keyName)
	}
	tc.c.Add(id, cacheElement{time.Now().Unix(), "", value})
}

//UnlinkName unlinks the name fro mthe ID
func (tc *TimedCache) UnlinkName(keyName string) {
	tc.Lock()
	val, ok := tc.nameMap[keyName]
	delete(tc.nameMap, keyName)
	tc.Unlock()
	if ok {
		v, ok := tc.c.Get(val)
		if ok {
			ce := v.(cacheElement)
			ce.keyName = ""
			tc.c.Add(val, ce)
		}

	}
}

//GetNameID gets the ID associated with a name
func (tc *TimedCache) GetNameID(keyName string) (nameID int64, ok bool) {
	tc.Lock()
	defer tc.Unlock()
	val, ok := tc.nameMap[keyName]
	return val, ok
}

//UnlinkNamePrefix removes all names with the given prefix. This is not a very efficient way to do it,
//since it loops through the entire cache unlinking names.  It locks queries by name for the entire time.
//it might be a better idea to just let names expire, since if a user is deleted, any database operation on the user
//will fail miserably. Writing to a stream belonging to a deleted user is still possible that way, but
//Redis has a cleaning process that runs periodically.
func (tc *TimedCache) UnlinkNamePrefix(namePrefix string) {
	tc.Lock()
	for key := range tc.nameMap {
		if strings.HasPrefix(key, namePrefix) {
			delete(tc.nameMap, key)
		}
	}
	tc.Unlock()
}

//GetByID gets an element if it is cached and not yet expired
func (tc *TimedCache) GetByID(key int64) (value interface{}, keyName string, ok bool) {
	v, ok := tc.c.Get(key)
	if !ok {
		return nil, "", false
	}
	ce := v.(cacheElement)
	if ce.loadTime+tc.expireTime <= time.Now().Unix() {
		//The element is expired - it shouldn't be in the cache
		tc.UnlinkName(ce.keyName)
		tc.c.Remove(key)
		return nil, "", false
	}
	return ce.element, ce.keyName, true
}

//GetByName gets an element if it is in the cache and not expired
func (tc *TimedCache) GetByName(name string) (value interface{}, ok bool) {
	key, ok := tc.GetNameID(name)
	if !ok {
		return nil, false
	}
	value, keyName, ok := tc.GetByID(key)
	if keyName != name { //If ok is false, keyName = ""
		tc.UnlinkName(name) //Make sure that this is a matching name and ID
		ok = false
	}
	return value, ok
}

//RemoveID the given key from the cache by ID
func (tc *TimedCache) RemoveID(key int64) {
	v, ok := tc.c.Get(key)
	if ok {
		tc.UnlinkName(v.(cacheElement).keyName)
	}
	tc.c.Remove(key)
}

//RemoveName deletes the cache element by its name
func (tc *TimedCache) RemoveName(name string) {
	id, ok := tc.GetNameID(name)
	tc.UnlinkName(name)
	if ok {
		tc.RemoveID(id)
	}
}

//PurgeNames clears all the names from the cache, but leaves the cache elements accessible by ID
func (tc *TimedCache) PurgeNames() {
	tc.Lock()
	tc.nameMap = make(map[string]int64)
	tc.Unlock()
}

//Purge clears all elements in the cache
func (tc *TimedCache) Purge() {
	tc.PurgeNames()
	tc.c.Purge()
}
