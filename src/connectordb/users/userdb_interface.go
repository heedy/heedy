/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

/** UserDatabase is a base interface for specifying various database
functionality.

It can be used directly by the SqlUserDatabase, which performs all queries
directly, or it can be wrapped to include caching or logging.

**/
type UserDatabase interface {
	// User/Device/Stream limits are in config. The UserDatabase does not have access to the config
	CreateDevice(dm *DeviceMaker) error
	CreateStream(sm *StreamMaker) error
	CreateUser(um *UserMaker) error
	DeleteDevice(Id int64) error
	DeleteStream(Id int64) error
	DeleteUser(UserID int64) error
	Login(Username, Password string) (*User, *Device, error)
	ReadAllUsers() ([]*User, error)
	ReadDeviceByAPIKey(Key string) (*Device, error)
	ReadDeviceByID(DeviceID int64) (*Device, error)
	ReadDeviceForUserByName(userid int64, devicename string) (*Device, error)
	ReadDevicesForUserID(UserID int64) ([]*Device, error)
	ReadStreamByDeviceIDAndName(DeviceID int64, streamName string) (*Stream, error)
	ReadStreamByID(StreamID int64) (*Stream, error)
	ReadStreamsByDevice(DeviceID int64) ([]*Stream, error)
	ReadStreamsByUser(UserID int64) ([]*Stream, error)
	ReadUserById(UserID int64) (*User, error)
	ReadUserByName(Name string) (*User, error)
	ReadUserOperatingDevice(user *User) (*Device, error)
	UpdateDevice(device *Device) error
	UpdateStream(stream *Stream) error
	UpdateUser(user *User) error

	// Returns the total number of users in the database
	CountUsers() (int64, error)
	CountDevices() (int64, error)
	CountStreams() (int64, error)

	// Clears the database of all data
	Clear()
}
