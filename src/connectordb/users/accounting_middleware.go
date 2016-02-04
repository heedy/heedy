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

func (userdb *AccountingMiddleware) CreateDevice(Name string, UserID int64, public bool, devicelimit int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.CreateDevice(Name, UserID, public, devicelimit)
}

func (userdb *AccountingMiddleware) CreateStream(Name, Type string, DeviceID, streamlimit int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.CreateStream(Name, Type, DeviceID, streamlimit)
}

func (userdb *AccountingMiddleware) CreateUser(Name, Email, Password, Permissions string, public bool, userlimit int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.CreateUser(Name, Email, Password, Permissions, public, userlimit)
}

func (userdb *AccountingMiddleware) DeleteDevice(Id int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.DeleteDevice(Id)
}

func (userdb *AccountingMiddleware) DeleteStream(Id int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.DeleteStream(Id)
}

func (userdb *AccountingMiddleware) DeleteUser(UserID int64) error {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.DeleteUser(UserID)
}

func (userdb *AccountingMiddleware) Login(Username, Password string) (*User, *Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.Login(Username, Password)
}

func (userdb *AccountingMiddleware) ReadAllUsers() ([]*User, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadAllUsers()
}

func (userdb *AccountingMiddleware) ReadDeviceByAPIKey(Key string) (*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDeviceByAPIKey(Key)
}

func (userdb *AccountingMiddleware) ReadDeviceByID(DeviceID int64) (*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDeviceByID(DeviceID)
}

func (userdb *AccountingMiddleware) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDeviceForUserByName(userid, devicename)
}

func (userdb *AccountingMiddleware) ReadDevicesForUserID(UserID int64) ([]*Device, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadDevicesForUserID(UserID)
}

func (userdb *AccountingMiddleware) ReadStreamByDeviceIDAndName(DeviceID int64, streamName string) (*Stream, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadStreamByDeviceIDAndName(DeviceID, streamName)
}

func (userdb *AccountingMiddleware) ReadStreamByID(StreamID int64) (*Stream, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadStreamByID(StreamID)
}

func (userdb *AccountingMiddleware) ReadStreamsByDevice(DeviceID int64) ([]*Stream, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadStreamsByDevice(DeviceID)
}

func (userdb *AccountingMiddleware) ReadUserById(UserID int64) (*User, error) {
	atomic.AddUint64(&userdb.databaseCalls, 1)
	return userdb.UserDatabase.ReadUserById(UserID)
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
