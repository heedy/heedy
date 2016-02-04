/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.

This is the identity userdb, it probably won't be used in production, but it
can be handy while building new userdatabases.
**/
package users

type IdentityMiddleware struct {
	UserDatabase // the parent
}

func (userdb *IdentityMiddleware) Clear() {
	userdb.UserDatabase.Clear()
}

func (userdb *IdentityMiddleware) CreateDevice(Name string, UserID int64, public bool, devicelimit int64) error {
	return userdb.UserDatabase.CreateDevice(Name, UserID, public, devicelimit)
}

func (userdb *IdentityMiddleware) CreateStream(Name, Type string, DeviceID, streamlimit int64) error {
	return userdb.UserDatabase.CreateStream(Name, Type, DeviceID, streamlimit)
}

func (userdb *IdentityMiddleware) CreateUser(Name, Email, Password, Permissions string, public bool, userlimit int64) error {
	return userdb.UserDatabase.CreateUser(Name, Email, Password, Permissions, public, userlimit)
}

func (userdb *IdentityMiddleware) DeleteDevice(Id int64) error {
	return userdb.UserDatabase.DeleteDevice(Id)
}

func (userdb *IdentityMiddleware) DeleteStream(Id int64) error {
	return userdb.UserDatabase.DeleteStream(Id)
}

func (userdb *IdentityMiddleware) DeleteUser(UserID int64) error {
	return userdb.UserDatabase.DeleteUser(UserID)
}

func (userdb *IdentityMiddleware) Login(Username, Password string) (*User, *Device, error) {
	return userdb.UserDatabase.Login(Username, Password)
}

func (userdb *IdentityMiddleware) ReadAllUsers() ([]*User, error) {
	return userdb.UserDatabase.ReadAllUsers()
}

func (userdb *IdentityMiddleware) ReadDeviceByAPIKey(Key string) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceByAPIKey(Key)
}

func (userdb *IdentityMiddleware) ReadDeviceByID(DeviceID int64) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceByID(DeviceID)
}

func (userdb *IdentityMiddleware) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	return userdb.UserDatabase.ReadDeviceForUserByName(userid, devicename)
}

func (userdb *IdentityMiddleware) ReadDevicesForUserID(UserID int64) ([]*Device, error) {
	return userdb.UserDatabase.ReadDevicesForUserID(UserID)
}

func (userdb *IdentityMiddleware) ReadStreamByDeviceIDAndName(DeviceID int64, streamName string) (*Stream, error) {
	return userdb.UserDatabase.ReadStreamByDeviceIDAndName(DeviceID, streamName)
}

func (userdb *IdentityMiddleware) ReadStreamByID(StreamID int64) (*Stream, error) {
	return userdb.UserDatabase.ReadStreamByID(StreamID)
}

func (userdb *IdentityMiddleware) ReadStreamsByDevice(DeviceID int64) ([]*Stream, error) {
	return userdb.UserDatabase.ReadStreamsByDevice(DeviceID)
}

func (userdb *IdentityMiddleware) ReadUserById(UserID int64) (*User, error) {
	return userdb.UserDatabase.ReadUserById(UserID)
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
