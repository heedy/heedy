package users

import (
	"fmt"

	"github.com/josephlewis42/multicache"
)

/** CacheMiddleware provides an in-memory locally safe cache for various
get commands.
**/
type CacheMiddleware struct {
	UserDatabase // the parent

	// The three caches for storing the things we need
	userCache   *multicache.Multicache
	deviceCache *multicache.Multicache
	streamCache *multicache.Multicache
}

/** Creates and instantiates a new Caching middleware with the given parent and
cache sizes. Returns an error if the cache sizes are invalid (0)
**/
func NewCacheMiddleware(parent UserDatabase, userCacheSize, deviceCacheSize, streamCacheSize uint64) (UserDatabase, error) {

	userCache, err := multicache.NewDefaultMulticache(userCacheSize)
	if err != nil {
		return nil, err
	}

	deviceCache, err := multicache.NewDefaultMulticache(deviceCacheSize)
	if err != nil {
		return nil, err
	}

	streamCache, err := multicache.NewDefaultMulticache(streamCacheSize)
	if err != nil {
		return nil, err
	}

	var cm = CacheMiddleware{parent, userCache, deviceCache, streamCache}

	return &cm, nil
}

// Removes a particular user and its dependents from the cache
func (userdb *CacheMiddleware) clearCachedUser(UserId int64) {

	// Grab the streams we're supposed to remove.
	streams, err := userdb.ReadStreamsByUser(UserId)

	// Something bad happened, dump everything
	if err != nil {
		userdb.userCache.Purge()
		userdb.deviceCache.Purge()
		userdb.streamCache.Purge()
		return
	}

	// Perform the removes
	userdb.userCache.Remove(fmt.Sprintf("id:%d", UserId))

	// Remove all devices with the proper UserId
	userdb.deviceCache.RemoveManyFunc(func(item interface{}) (shouldRemove bool) {
		dev := item.(Device)
		return dev.UserId == UserId
	})

	// Change our streams into a set for O(n) search.
	streamIds := map[int64]bool{}
	for _, stream := range streams {
		streamIds[stream.StreamId] = true
	}

	userdb.streamCache.RemoveManyFunc(func(item interface{}) bool {
		stream := item.(Stream)
		_, ok := streamIds[stream.StreamId]
		return ok
	})
}

// Removes a device from the caches along with all its children streams
func (userdb *CacheMiddleware) clearCachedDevice(DeviceId int64) {
	userdb.deviceCache.Remove(fmt.Sprintf("id:%d", DeviceId))

	// Now remove all streams that are children
	userdb.streamCache.RemoveManyFunc(func(item interface{}) (shouldRemove bool) {
		stream := item.(Stream)
		return stream.DeviceId == DeviceId
	})
}

// Removes a stream with the given id from the cache
func (userdb *CacheMiddleware) clearCachedStream(StreamId int64) {
	userdb.streamCache.Remove(fmt.Sprintf("id:%d", StreamId))
}

func (userdb *CacheMiddleware) cacheUser(user *User, err error) {
	if err != nil || user == nil {
		return
	}

	cacheable := *user

	userdb.userCache.AddMany(cacheable,
		fmt.Sprintf("id:%d", user.UserId),
		fmt.Sprintf("name:%s", user.Name))
}

func (userdb *CacheMiddleware) cacheStream(stream *Stream, err error) {
	if err != nil || stream == nil {
		return
	}

	cacheable := *stream

	userdb.streamCache.AddMany(cacheable,
		fmt.Sprintf("id:%d", stream.StreamId),
		fmt.Sprintf("dev:%dname:%s", stream.DeviceId, stream.Name))
}

func (userdb *CacheMiddleware) cacheDevice(dev *Device, err error) {
	if err != nil || dev == nil {
		return
	}

	cacheable := *dev

	userdb.deviceCache.AddMany(cacheable,
		fmt.Sprintf("id:%d", dev.DeviceId),
		fmt.Sprintf("usr:%dname:%s", dev.UserId, dev.Name),
		fmt.Sprintf("apikey:%s", dev.ApiKey))
}

func (userdb *CacheMiddleware) readUser(key string) (user User, ok bool) {

	tmp, ok := userdb.userCache.Get(key)
	if !ok {
		return User{}, ok
	}

	return tmp.(User), ok
}

func (userdb *CacheMiddleware) readStream(key string) (stream Stream, ok bool) {

	tmp, ok := userdb.streamCache.Get(key)
	if !ok {
		return Stream{}, ok
	}

	return tmp.(Stream), ok
}

func (userdb *CacheMiddleware) readDevice(key string) (dev Device, ok bool) {

	tmp, ok := userdb.deviceCache.Get(key)
	if !ok {
		return Device{}, ok
	}

	return tmp.(Device), ok
}

func (userdb *CacheMiddleware) DeleteDevice(Id int64) error {
	userdb.clearCachedDevice(Id)

	return userdb.UserDatabase.DeleteDevice(Id)
}

func (userdb *CacheMiddleware) DeleteStream(Id int64) error {
	userdb.clearCachedStream(Id)

	return userdb.UserDatabase.DeleteStream(Id)
}

func (userdb *CacheMiddleware) DeleteUser(UserId int64) error {
	// Do this first since we need the db in the same state
	userdb.clearCachedUser(UserId)

	return userdb.UserDatabase.DeleteUser(UserId)
}

func (userdb *CacheMiddleware) Login(Username, Password string) (*User, *Device, error) {
	user, dev, err := userdb.UserDatabase.Login(Username, Password)

	userdb.cacheUser(user, err)
	userdb.cacheDevice(dev, err)

	return user, dev, err
}

func (userdb *CacheMiddleware) ReadAllUsers() ([]User, error) {
	return userdb.UserDatabase.ReadAllUsers()
}

func (userdb *CacheMiddleware) ReadDeviceByApiKey(Key string) (*Device, error) {
	cacheDev, ok := userdb.readDevice("api:" + Key)
	if ok {
		return &cacheDev, nil
	}

	dev, err := userdb.UserDatabase.ReadDeviceByApiKey(Key)

	userdb.cacheDevice(dev, err)

	return dev, err
}

func (userdb *CacheMiddleware) ReadDeviceById(DeviceId int64) (*Device, error) {
	cacheDev, ok := userdb.readDevice(fmt.Sprintf("id:%d", DeviceId))
	if ok {
		return &cacheDev, nil
	}

	dev, err := userdb.UserDatabase.ReadDeviceById(DeviceId)

	userdb.cacheDevice(dev, err)

	return dev, err
}

func (userdb *CacheMiddleware) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	cacheDev, ok := userdb.readDevice(fmt.Sprintf("usr:%dname:%s", userid, devicename))
	if ok {
		return &cacheDev, nil
	}

	dev, err := userdb.UserDatabase.ReadDeviceForUserByName(userid, devicename)

	userdb.cacheDevice(dev, err)

	return dev, err
}

func (userdb *CacheMiddleware) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	cached, ok := userdb.readStream(fmt.Sprintf("dev:%dname:%s", DeviceId, streamName))
	if ok {
		return &cached, nil
	}

	stream, err := userdb.UserDatabase.ReadStreamByDeviceIdAndName(DeviceId, streamName)

	userdb.cacheStream(stream, err)

	return stream, err
}

func (userdb *CacheMiddleware) ReadStreamById(StreamId int64) (*Stream, error) {
	cacheStream, ok := userdb.readStream(fmt.Sprintf("id:%d", StreamId))
	if ok {
		return &cacheStream, nil
	}

	stream, err := userdb.UserDatabase.ReadStreamById(StreamId)

	userdb.cacheStream(stream, err)

	return stream, err
}

func (userdb *CacheMiddleware) ReadUserById(UserId int64) (*User, error) {
	cacheUser, ok := userdb.readUser(fmt.Sprintf("id:%d", UserId))
	if ok {
		return &cacheUser, nil
	}

	user, err := userdb.UserDatabase.ReadUserById(UserId)

	userdb.cacheUser(user, err)

	return user, err
}

func (userdb *CacheMiddleware) ReadUserByName(Name string) (*User, error) {
	cacheUser, ok := userdb.readUser(fmt.Sprintf("name:%s", Name))
	if ok {
		return &cacheUser, nil
	}

	user, err := userdb.UserDatabase.ReadUserByName(Name)

	userdb.cacheUser(user, err)

	return user, err
}

func (userdb *CacheMiddleware) ReadUserOperatingDevice(user *User) (*Device, error) {
	if user == nil {
		return nil, InvalidPointerError
	}

	cacheDev, ok := userdb.readDevice(fmt.Sprintf("usr:%dname:user", user.UserId))
	if ok {
		return &cacheDev, nil
	}

	return userdb.UserDatabase.ReadUserOperatingDevice(user)
}

func (userdb *CacheMiddleware) UpdateDevice(device *Device) error {
	if device == nil {
		return InvalidPointerError
	}

	err := userdb.UserDatabase.UpdateDevice(device)
	userdb.clearCachedDevice(device.DeviceId)
	return err
}

func (userdb *CacheMiddleware) UpdateStream(stream *Stream) error {
	if stream == nil {
		return InvalidPointerError
	}

	err := userdb.UserDatabase.UpdateStream(stream)
	userdb.clearCachedStream(stream.StreamId)
	return err
}

func (userdb *CacheMiddleware) UpdateUser(user *User) error {
	if user == nil {
		return InvalidPointerError
	}

	// Do this first since the db should still be in the same state
	userdb.clearCachedUser(user.UserId)
	err := userdb.UserDatabase.UpdateUser(user)
	// Do this again since permissions can change when updating a user.
	userdb.clearCachedUser(user.UserId)

	return err
}
