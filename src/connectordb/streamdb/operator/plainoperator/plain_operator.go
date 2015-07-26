package plainoperator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator/messenger"
	"connectordb/streamdb/users"
	"errors"
)

const (
	PlainOperatorName = " ADMIN "
)

var (
	ErrAdmin = errors.New("An administrative operator has no user or device")
)

/**

The basic database access operator, overload anything here that you need to get
functionality right.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>

All Rights Reserved

**/

// This operator is very insecure but very fast, good for embedded environments
// where all you care about is speed and can trust all your users
type PlainOperator struct {
	Userdb users.UserDatabase     // SqlUserDatabase holds the methods needed to CRUD users/devices/streams
	ds     *datastream.DataStream // datastream holds methods for inserting datapoints into streams
	msg    *messenger.Messenger   // messenger is a connection to the messaging client
}

//Name here is a special one meaning that it is the database administration operator
// It is not a valid username
func (db *PlainOperator) Name() string {
	return PlainOperatorName
}

//User returns the current user
func (db *PlainOperator) User() (usr *users.User, err error) {
	return nil, ErrAdmin
}

func (db *PlainOperator) Device() (*users.Device, error) {
	return nil, ErrAdmin
}

func (o *PlainOperator) CreateUser(username, email, password string) error {
	return o.Userdb.CreateUser(username, email, password)
}

func (o *PlainOperator) ReadAllUsers() ([]users.User, error) {
	return o.Userdb.ReadAllUsers()
}

func (o *PlainOperator) ReadUser(username string) (*users.User, error) {
	return o.Userdb.ReadUserByName(username)
}

func (o *PlainOperator) ReadUserByID(userID int64) (*users.User, error) {
	return o.Userdb.ReadUserById(userID)
}

func (o *PlainOperator) UpdateUser(modifieduser *users.User) error {
	return o.Userdb.UpdateUser(modifieduser)
}
func (o *PlainOperator) DeleteUserByID(userID int64) error {
	// Users are going to be GC'd from redis in the future - but we currently don't have that implemented,
	// so manually delete all the devices from redis if user delete succeeds
	dev, err := o.ReadAllDevicesByUserID(userID)
	if err != nil {
		return err
	}

	err = o.Userdb.DeleteUser(userID)

	if err == nil {
		for i := 0; i < len(dev); i++ {
			o.ds.DeleteDevice(dev[i].DeviceId)
		}
	}
	return err
}

func (o *PlainOperator) ReadAllDevicesByUserID(userID int64) ([]users.Device, error) {
	return o.Userdb.ReadDevicesForUserId(userID)
}

func (o *PlainOperator) CreateDeviceByUserID(userID int64, deviceName string) error {
	return o.Userdb.CreateDevice(deviceName, userID)
}

func (o *PlainOperator) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	return o.Userdb.ReadDeviceById(deviceID)
}

func (o *PlainOperator) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	return o.Userdb.ReadDeviceForUserByName(userID, devicename)
}

func (o *PlainOperator) UpdateDevice(modifieddevice *users.Device) error {
	return o.Userdb.UpdateDevice(modifieddevice)
}

func (o *PlainOperator) DeleteDeviceByID(deviceID int64) error {
	err := o.Userdb.DeleteDevice(deviceID)
	if err == nil {
		err = o.ds.DeleteDevice(deviceID)
	}
	return err
}

func (o *PlainOperator) CountUsers() (uint64, error) {
	return o.Userdb.CountUsers()
}

func (o *PlainOperator) CountDevices() (uint64, error) {
	return o.Userdb.CountUsers()
}

func (o *PlainOperator) CountStreams() (uint64, error) {
	return o.Userdb.CountUsers()
}

func (o *PlainOperator) ReadAllStreamsByDeviceID(deviceID int64) ([]users.Stream, error) {
	return o.Userdb.ReadStreamsByDevice(deviceID)
}

func (o *PlainOperator) CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error {
	return o.Userdb.CreateStream(streamname, jsonschema, deviceID)
}

//ReadStreamByID reads a stream using a stream's ID
func (o *PlainOperator) ReadStreamByID(streamID int64) (*users.Stream, error) {
	return o.Userdb.ReadStreamById(streamID)
}

//ReadStreamByDeviceID reads a stream given its name and the ID of its parent device
func (o *PlainOperator) ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error) {
	return o.Userdb.ReadStreamByDeviceIdAndName(deviceID, streamname)
}

//UpdateStream updates the stream. BUG(daniel) the function currently does not give an error
//if someone attempts to update the schema (which is an illegal operation anyways)
func (o *PlainOperator) UpdateStream(modifiedstream *users.Stream) error {
	strm, err := o.ReadStreamByID(modifiedstream.StreamId)
	if err != nil {
		return err
	}

	err = o.Userdb.UpdateStream(modifiedstream)

	if err == nil && strm.Downlink == true && modifiedstream.Downlink == false {
		//There was a downlink here. Since the downlink was removed, we delete the associated
		//downlink substream
		o.DeleteStreamByID(strm.StreamId, "downlink")
	}

	return err
}

//DeleteStreamByID deletes the stream using ID
func (o *PlainOperator) DeleteStreamByID(streamID int64, substream string) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err //Workaround #81
	}

	if substream != "" {
		//We just delete the substream
		err = o.ds.DeleteSubstream(strm.DeviceId, strm.StreamId, substream)
	} else {
		err = o.Userdb.DeleteStream(streamID)
		if err == nil {
			err = o.ds.DeleteStream(strm.DeviceId, strm.StreamId)
		}
	}
	return err

}
