package users

/** This is the identity userdb, it probably won't be used in production, but it
can be handy while building new userdatabases.

**/
type IdentityUserdb struct {
	UserDatabase // the parent
}

func (userdb *IdentityUserdb) CreateDevice(Name string, UserId int64) error {
	return userdb.UserDatabase.CreateDevice(Name, UserId)
}

func (userdb *IdentityUserdb) CreateStream(Name, Type string, DeviceId int64) error {
	return userdb.UserDatabase.CreateStream(Name, Type, DeviceId)
}

func (userdb *IdentityUserdb) CreateUser(Name, Email, Password string) error {
	return userdb.UserDatabase.CreateUser(Name, Email, Password)
}

func (userdb *IdentityUserdb) DeleteDevice(Id int64) error {
	return userdb.UserDatabase.DeleteDevice(Id)
}

func (userdb *IdentityUserdb) DeleteStream(Id int64) error {
	return userdb.UserDatabase.DeleteStream(Id)
}

func (userdb *IdentityUserdb) DeleteUser(UserId int64) error {
	return userdb.UserDatabase.DeleteUser(UserId)
}

func (userdb *IdentityUserdb) Login(Username, Password string) (*User, *Device, error) {
	return userdb.UserDatabase.Login(Username, Password)
}

func (userdb *IdentityUserdb) ReadAllUsers() ([]User, error) {
	return userdb.UserDatabase.ReadAllUsers()
}

func (userdb *IdentityUserdb) ReadDeviceByApiKey(Key string) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceByApiKey(Key)
}

func (userdb *IdentityUserdb) ReadDeviceById(DeviceId int64) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceById(DeviceId)
}

func (userdb *IdentityUserdb) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceForUserByName(userid, devicename)
}

func (userdb *IdentityUserdb) ReadDevicesForUserId(UserId int64) ([]Device, error) {
	return userdb.UserDatabase.ReadDevicesForUserId(UserId)
}

func (userdb *IdentityUserdb) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	return userdb.UserDatabase.ReadStreamByDeviceIdAndName(DeviceId, streamName)
}

func (userdb *IdentityUserdb) ReadStreamById(StreamId int64) (*Stream, error) {
	return userdb.UserDatabase.ReadStreamById(StreamId)
}

func (userdb *IdentityUserdb) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	return userdb.UserDatabase.ReadStreamsByDevice(DeviceId)
}

func (userdb *IdentityUserdb) ReadUserById(UserId int64) (*User, error) {
	return userdb.UserDatabase.ReadUserById(UserId)
}

func (userdb *IdentityUserdb) ReadUserByName(Name string) (*User, error) {
	return userdb.UserDatabase.ReadUserByName(Name)
}

func (userdb *IdentityUserdb) ReadUserOperatingDevice(user *User) (*Device, error) {
	return userdb.UserDatabase.ReadUserOperatingDevice(user)
}

func (userdb *IdentityUserdb) UpdateDevice(device *Device) error {
	return userdb.UserDatabase.UpdateDevice(device)
}

func (userdb *IdentityUserdb) UpdateStream(stream *Stream) error {
	return userdb.UserDatabase.UpdateStream(stream)
}

func (userdb *IdentityUserdb) UpdateUser(user *User) error {
	return userdb.UserDatabase.UpdateUser(user)
}
