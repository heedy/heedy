/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.

Provides the ability to count the number of users/devices/streams in the database
**/
package users

func (userdb *SqlUserDatabase) CountUsers() (int64, error) {
	var output int64
	err := userdb.Get(&output, "SELECT COUNT(UserID) FROM Users;")
	return output, err
}

func (userdb *SqlUserDatabase) CountStreams() (int64, error) {
	var output int64
	err := userdb.Get(&output, "SELECT COUNT(StreamID) FROM Streams;")
	return output, err
}

func (userdb *SqlUserDatabase) CountDevices() (int64, error) {
	var output int64
	err := userdb.Get(&output, "SELECT COUNT(DeviceID) FROM Devices;")
	return output, err
}

func (userdb *SqlUserDatabase) CountStreamsForDevice(DeviceID int64) (int64, error) {
	var output int64
	err := userdb.Get(&output, "SELECT COUNT(StreamID) FROM Streams WHERE DeviceID = ?;", DeviceID)
	return output, err
}

func (userdb *SqlUserDatabase) CountDevicesForUser(UserID int64) (int64, error) {
	var output int64
	err := userdb.Get(&output, "SELECT COUNT(DeviceID) FROM Devices WHERE UserID = ?;", UserID)
	return output, err
}
