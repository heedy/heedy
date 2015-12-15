/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.

This is the a middleware that counts the number of calls to the database.
**/
package users

import "sync/atomic"

type AccountingMiddleware struct {
	UserDatabase // the parent

	// Number of database calls since last Reset()
	databaseCalls uint64
}

func (userdb *AccountingMiddleware) GetNumberOfCalls() uint64 {
	return atomic.LoadUint64(&userdb.databaseCalls)
}

func (userdb *AccountingMiddleware) CreateDevice(Name string, UserId int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.CreateDevice(Name, UserId)
}

func (userdb *AccountingMiddleware) CreateStream(Name, Type string, DeviceId int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.CreateStream(Name, Type, DeviceId)
}

func (userdb *AccountingMiddleware) CreateUser(Name, Email, Password string) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.CreateUser(Name, Email, Password)
}

func (userdb *AccountingMiddleware) DeleteDevice(Id int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.DeleteDevice(Id)
}

func (userdb *AccountingMiddleware) DeleteStream(Id int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.DeleteStream(Id)
}

func (userdb *AccountingMiddleware) DeleteUser(UserId int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.DeleteUser(UserId)
}

func (userdb *AccountingMiddleware) Login(Username, Password string) (*User, *Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.Login(Username, Password)
}

func (userdb *AccountingMiddleware) ReadAllUsers() ([]User, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadAllUsers()
}

func (userdb *AccountingMiddleware) ReadDeviceByApiKey(Key string) (*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDeviceByApiKey(Key)
}

func (userdb *AccountingMiddleware) ReadDeviceById(DeviceId int64) (*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDeviceById(DeviceId)
}

func (userdb *AccountingMiddleware) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDeviceForUserByName(userid, devicename)
}

func (userdb *AccountingMiddleware) ReadDevicesForUserId(UserId int64) ([]Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDevicesForUserId(UserId)
}

func (userdb *AccountingMiddleware) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadStreamByDeviceIdAndName(DeviceId, streamName)
}

func (userdb *AccountingMiddleware) ReadStreamById(StreamId int64) (*Stream, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadStreamById(StreamId)
}

func (userdb *AccountingMiddleware) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadStreamsByDevice(DeviceId)
}

func (userdb *AccountingMiddleware) ReadUserById(UserId int64) (*User, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadUserById(UserId)
}

func (userdb *AccountingMiddleware) ReadUserByName(Name string) (*User, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadUserByName(Name)
}

func (userdb *AccountingMiddleware) ReadUserOperatingDevice(user *User) (*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadUserOperatingDevice(user)
}

func (userdb *AccountingMiddleware) UpdateDevice(device *Device) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.UpdateDevice(device)
}

func (userdb *AccountingMiddleware) UpdateStream(stream *Stream) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.UpdateStream(stream)
}

func (userdb *AccountingMiddleware) UpdateUser(user *User) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.UpdateUser(user)
}
