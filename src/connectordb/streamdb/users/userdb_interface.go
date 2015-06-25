package users

/** UserDatabase is a base interface for specifying various database
functionality.

It can be used directly by the SqlUserDatabase, which performs all queries
directly, or it can be wrapped to include caching or logging.

**/
type UserDatabase interface {
	CreateDevice(Name string, UserId int64) error
	CreateStream(Name, Type string, DeviceId int64) error
	CreateUser(Name, Email, Password string) error
	DeleteDevice(Id int64) error
	DeleteStream(Id int64) error
	DeleteUser(UserId int64) error
	Login(Username, Password string) (*User, *Device, error)
	ReadAllUsers() ([]User, error)
	ReadDeviceByApiKey(Key string) (*Device, error)
	ReadDeviceById(DeviceId int64) (*Device, error)
	ReadDeviceForUserByName(userid int64, devicename string) (*Device, error)
	ReadDevicesForUserId(UserId int64) ([]Device, error)
	ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error)
	ReadStreamById(StreamId int64) (*Stream, error)
	ReadStreamsByDevice(DeviceId int64) ([]Stream, error)
	ReadUserById(UserId int64) (*User, error)
	ReadUserByName(Name string) (*User, error)
	ReadUserOperatingDevice(user *User) (*Device, error)
	UpdateDevice(device *Device) error
	UpdateStream(stream *Stream) error
	UpdateUser(user *User) error
}
