package users

/** UserDatabase is a base interface for specifying various database
functionality.

It can be used directly by the SqlUserDatabase, which performs all queries
directly, or it can be wrapped to include caching or logging.

**/
type CacheMiddleware struct {
	UserDatabase // the parent

}

func (userdb *CacheMiddleware) clearCaches() {
	// TODO implement me
}

func (userdb *CacheMiddleware) cacheDevice(dev *Device, err error) {
	if err != nil || dev == nil {
		return
	}

	// TODO implement me
}

func (userdb *CacheMiddleware) cacheUser(user *User, err error) {
	if err != nil || user == nil {
		return
	}

	// TODO implement me
}

func (userdb *CacheMiddleware) cacheStream(stream *Stream, err error) {
	if err != nil || stream == nil {
		return
	}

	// TODO implement me
}

func (userdb *CacheMiddleware) CreateDevice(Name string, UserId int64) error {
	err := userdb.UserDatabase.CreateDevice(Name, UserId)
	return err
}

func (userdb *CacheMiddleware) CreateStream(Name, Type string, DeviceId int64) error {
	err := userdb.UserDatabase.CreateStream(Name, Type, DeviceId)
	return err
}

func (userdb *CacheMiddleware) DeleteDevice(Id int64) error {
	err := userdb.UserDatabase.DeleteDevice(Id)
	// As for now, we have no idea what percentage of requests will be deletes,
	// the assumption is that they will be very small, which seems reasonable.
	// as such, it isn't worth making the code "smarter" due to the inherently
	// higher complexity and potential side-effects.
	userdb.clearCaches()
	return err
}

func (userdb *CacheMiddleware) DeleteStream(Id int64) error {
	err := userdb.UserDatabase.DeleteStream(Id)
	// As for now, we have no idea what percentage of requests will be deletes,
	// the assumption is that they will be very small, which seems reasonable.
	// as such, it isn't worth making the code "smarter" due to the inherently
	// higher complexity and potential side-effects.
	userdb.clearCaches()
	return err
}

func (userdb *CacheMiddleware) DeleteUser(UserId int64) error {
	err := userdb.UserDatabase.DeleteUser(UserId)
	// As for now, we have no idea what percentage of requests will be deletes,
	// the assumption is that they will be very small, which seems reasonable.
	// as such, it isn't worth making the code "smarter" due to the inherently
	// higher complexity and potential side-effects.
	userdb.clearCaches()
	return err
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
	dev, err := userdb.UserDatabase.ReadDeviceByApiKey(Key)

	userdb.cacheDevice(dev, err)

	return dev, err
}

func (userdb *CacheMiddleware) ReadDeviceById(DeviceId int64) (*Device, error) {
	dev, err := userdb.UserDatabase.ReadDeviceById(DeviceId)

	userdb.cacheDevice(dev, err)

	return dev, err
}

func (userdb *CacheMiddleware) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	dev, err := userdb.UserDatabase.ReadDeviceForUserByName(userid, devicename)

	userdb.cacheDevice(dev, err)

	return dev, err
}

func (userdb *CacheMiddleware) ReadDevicesForUserId(UserId int64) ([]Device, error) {
	return userdb.UserDatabase.ReadDevicesForUserId(UserId)
}

func (userdb *CacheMiddleware) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	stream, err := userdb.UserDatabase.ReadStreamByDeviceIdAndName(DeviceId, streamName)

	userdb.cacheStream(stream, err)

	return stream, err
}

func (userdb *CacheMiddleware) ReadStreamById(StreamId int64) (*Stream, error) {
	stream, err := userdb.UserDatabase.ReadStreamById(StreamId)

	userdb.cacheStream(stream, err)

	return stream, err
}

func (userdb *CacheMiddleware) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	return userdb.UserDatabase.ReadStreamsByDevice(DeviceId)
}

func (userdb *CacheMiddleware) ReadUserById(UserId int64) (*User, error) {
	user, err := userdb.UserDatabase.ReadUserById(UserId)

	userdb.cacheUser(user, err)

	return user, err
}

func (userdb *CacheMiddleware) ReadUserByName(Name string) (*User, error) {
	user, err := userdb.UserDatabase.ReadUserByName(Name)

	userdb.cacheUser(user, err)

	return user, err
}

func (userdb *CacheMiddleware) ReadUserOperatingDevice(user *User) (*Device, error) {
	return userdb.UserDatabase.ReadUserOperatingDevice(user)
}

func (userdb *CacheMiddleware) UpdateDevice(device *Device) error {
	err := userdb.UserDatabase.UpdateDevice(device)
	// As for now, we have no idea what percentage of requests updates will be,
	// the assumption is that they will be very small, which seems reasonable.
	// as such, it isn't worth making the code "smarter" due to the inherently
	// higher complexity and potential side-effects.
	userdb.clearCaches()
	return err
}

func (userdb *CacheMiddleware) UpdateStream(stream *Stream) error {
	err := userdb.UserDatabase.UpdateStream(stream)
	// As for now, we have no idea what percentage of requests updates will be,
	// the assumption is that they will be very small, which seems reasonable.
	// as such, it isn't worth making the code "smarter" due to the inherently
	// higher complexity and potential side-effects.
	userdb.clearCaches()
	return err
}

func (userdb *CacheMiddleware) UpdateUser(user *User) error {
	err := userdb.UserDatabase.UpdateUser(user)
	// As for now, we have no idea what percentage of requests updates will be,
	// the assumption is that they will be very small, which seems reasonable.
	// as such, it isn't worth making the code "smarter" due to the inherently
	// higher complexity and potential side-effects.
	userdb.clearCaches()
	return err
}
