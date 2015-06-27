package users

/** This is the identity userdb, it probably won't be used in production, but it
can be handy while building new userdatabases.
**/

type IdentityMiddleware struct {
	UserDatabase // the parent
}

func (userdb *IdentityMiddleware) CreateDevice(Name string, UserId int64) error {
	return userdb.UserDatabase.CreateDevice(Name, UserId)
}

func (userdb *IdentityMiddleware) CreateStream(Name, Type string, DeviceId int64) error {
	return userdb.UserDatabase.CreateStream(Name, Type, DeviceId)
}

func (userdb *IdentityMiddleware) CreateUser(Name, Email, Password string) error {
	return userdb.UserDatabase.CreateUser(Name, Email, Password)
}

func (userdb *IdentityMiddleware) DeleteDevice(Id int64) error {
	return userdb.UserDatabase.DeleteDevice(Id)
}

func (userdb *IdentityMiddleware) DeleteStream(Id int64) error {
	return userdb.UserDatabase.DeleteStream(Id)
}

func (userdb *IdentityMiddleware) DeleteUser(UserId int64) error {
	return userdb.UserDatabase.DeleteUser(UserId)
}

func (userdb *IdentityMiddleware) Login(Username, Password string) (*User, *Device, error) {
	return userdb.UserDatabase.Login(Username, Password)
}

func (userdb *IdentityMiddleware) ReadAllUsers() ([]User, error) {
	return userdb.UserDatabase.ReadAllUsers()
}

func (userdb *IdentityMiddleware) ReadDeviceByApiKey(Key string) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceByApiKey(Key)
}

func (userdb *IdentityMiddleware) ReadDeviceById(DeviceId int64) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceById(DeviceId)
}

func (userdb *IdentityMiddleware) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceForUserByName(userid, devicename)
}

func (userdb *IdentityMiddleware) ReadDevicesForUserId(UserId int64) ([]Device, error) {
	return userdb.UserDatabase.ReadDevicesForUserId(UserId)
}

func (userdb *IdentityMiddleware) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	return userdb.UserDatabase.ReadStreamByDeviceIdAndName(DeviceId, streamName)
}

func (userdb *IdentityMiddleware) ReadStreamById(StreamId int64) (*Stream, error) {
	return userdb.UserDatabase.ReadStreamById(StreamId)
}

func (userdb *IdentityMiddleware) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	return userdb.UserDatabase.ReadStreamsByDevice(DeviceId)
}

func (userdb *IdentityMiddleware) ReadUserById(UserId int64) (*User, error) {
	return userdb.UserDatabase.ReadUserById(UserId)
}

func (userdb *IdentityMiddleware) ReadUserByName(Name string) (*User, error) {
	return userdb.UserDatabase.ReadUserByName(Name)
}

func (userdb *IdentityMiddleware) ReadUserOperatingDevice(user *User) (*Device, error) {
	return userdb.UserDatabase.ReadUserOperatingDevice(user)
}

func (userdb *IdentityMiddleware) UpdateDevice(device *Device) error {
	return userdb.UserDatabase.UpdateDevice(device)
}

func (userdb *IdentityMiddleware) UpdateStream(stream *Stream) error {
	return userdb.UserDatabase.UpdateStream(stream)
}

func (userdb *IdentityMiddleware) UpdateUser(user *User) error {
	return userdb.UserDatabase.UpdateUser(user)
}
