/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.

This is the error userdb, it probably won't be used in production, but it
can be handy while doing testing to ensure everything returns an error.
**/
package users

import "errors"

var (
	ErrorUserdbError = errors.New("Error Middleware Error")
)

type ErrorUserdb struct {
}

func (userdb *ErrorUserdb) Clear() {
}

func (userdb *ErrorUserdb) CreateDevice(Name string, UserID int64, public bool, devicelimit int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) CreateStream(Name, Type string, DeviceID, streamlimit int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) CreateUser(Name, Email, Password, Permissions string, Public bool, userlimit int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) DeleteDevice(Id int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) DeleteStream(Id int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) DeleteUser(UserID int64) error {
	return ErrorUserdbError
}

func (userdb *ErrorUserdb) Login(Username, Password string) (*User, *Device, error) {
	return nil, nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadAllUsers() ([]*User, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDeviceByAPIKey(Key string) (*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDeviceByID(DeviceID int64) (*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadDevicesForUserID(UserID int64) ([]*Device, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadStreamByDeviceIDAndName(DeviceID int64, streamName string) (*Stream, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadStreamByID(StreamID int64) (*Stream, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadStreamsByDevice(DeviceID int64) ([]*Stream, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadStreamsByUser(UserID int64) ([]*Stream, error) {
	return nil, ErrorUserdbError
}

func (userdb *ErrorUserdb) ReadUserById(UserID int64) (*User, error) {
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

func (userdb *ErrorUserdb) CountUsers() (int64, error) {
	return 1, ErrorUserdbError
}

func (userdb *ErrorUserdb) CountStreams() (int64, error) {
	return 1, ErrorUserdbError
}

func (userdb *ErrorUserdb) CountDevices() (int64, error) {
	return 1, ErrorUserdbError
}
