package users

import "errors"

/** This is the error userdb, it probably won't be used in production, but it
can be handy while doing testing to ensure everything returns an error.
**/

var (
	ErrorUserdbError = errors.New("Error Middleware Error")
)

type ErrorUserdb struct {
}

func (userdb *ErrorUserdb) CreateDevice(Name string, UserId int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) CreateStream(Name, Type string, DeviceId int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) CreateUser(Name, Email, Password string) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) DeleteDevice(Id int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) DeleteStream(Id int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) DeleteUser(UserId int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) Login(Username, Password string) (*User, *Device, error) {
	return nil, nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadAllUsers() ([]User, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDeviceByApiKey(Key string) (*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDeviceById(DeviceId int64) (*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDevicesForUserId(UserId int64) ([]Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadStreamById(StreamId int64) (*Stream, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadUserById(UserId int64) (*User, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadUserByName(Name string) (*User, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadUserOperatingDevice(user *User) (*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) UpdateDevice(device *Device) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) UpdateStream(stream *Stream) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) UpdateUser(user *User) error {
	return ErrorUserdbError
}
