package operator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator/adminoperator"
	"connectordb/streamdb/operator/messenger"
	"connectordb/streamdb/operator/plainoperator"
	"connectordb/streamdb/users"
)

type Database interface {
	GetUserDatabase() users.UserDatabase
	GetDatastream() *datastream.DataStream
	GetMessenger() *messenger.Messenger
}

func newPlainOperator(db *Database) PlainOperator {
	return plainoperator.NewPlainOperator(db.GetUserDatabase(), db.GetDatastream(), db.GetMessenger())

}

func NewPlainOperator(db *Database) Operator {
	return newPlainOperator(db)
}

func newAdminOperator(db *Database) AdminOperator {
	return adminoperator.AdminOperator{NewPlainOperator(udb, ds, msg)}
}

func NewAdminOperator(db *Database) Operator {
	return newAdminOperator(db)
}

func DeviceLoginOperator(db *Database, username, password string) (Operator, error) {
	baseOperator = newPlainOperator(db)
	dev, err := operator.ReadDevice(devicepath)

	if err != nil || dev.ApiKey != apikey {
		return operator.Operator{}, authoperator.ErrPermissions // Don't leak whether the device exists
	}
	return authoperator.NewAuthOperator(db, dev.DeviceId)
}

/**
//DeviceLoginOperator returns the operator associated with the given API key
func (db *Database) DeviceLoginOperator(devicepath, apikey string) (operator.Operator, error) {
	dev, err := db.ReadDevice(devicepath)
	if err != nil || dev.ApiKey != apikey {
		return operator.Operator{}, authoperator.ErrPermissions //Don't leak whether the device exists
	}
	return authoperator.NewAuthOperator(db, dev.DeviceId)
}


// LoginOperator logs in as a user or device, depending on which is passed in
func (db *Database) LoginOperator(path, password string) (operator.Operator, error) {
	switch strings.Count(path, "/") {
	default:
		return operator.Operator{}, operator.ErrBadPath
	case 1:
		return db.DeviceLoginOperator(path, password)
	case 0:
		return db.UserLoginOperator(path, password)
	}
}

//Operator gets the operator by usr or device name
func (db *Database) GetOperator(path string) (operator.Operator, error) {
	switch strings.Count(path, "/") {
	default:
		return operator.Operator{}, operator.ErrBadPath
	case 0:
		path += "/user"
	case 1:
		//Do nothing for this case
	}
	dev, err := db.ReadDevice(path)
	if err != nil {
		return operator.Operator{}, err //We use dev.Name, so must return error earlier
	}
	return authoperator.NewAuthOperator(db, dev.DeviceId)
}

//DeviceOperator returns the operator for the given device ID
func (db *Database) DeviceOperator(deviceID int64) (operator.Operator, error) {
	return authoperator.NewAuthOperator(db, deviceID)
}

/**
//NewAuthOperator creates a new authenticated operator,
func NewAuthOperator(db operator.BaseOperatorInterface, deviceID int64) (operator.PlainOperator, error) {
	dev, err := db.ReadDeviceByID(deviceID)
	if err != nil {
		return operator.Operator{}, err
	}
	usr, err := db.ReadUserByID(dev.UserId)
	if err != nil {
		return operator.Operator{}, err
	}

	userlogID, err := getUserLogStream(db, usr.UserId)
	if err != nil {
		return operator.Operator{}, err
	}

	return operator.{&AuthOperator{db, usr.Name + "/" + dev.Name, dev.DeviceId, userlogID}}, nil
}

//Returns the stream ID of the user log stream (and tries to create it if the stream does not exist)
func getUserLogStream(db operator.BaseOperatorInterface, userID int64) (streamID int64, err error) {
	o := operator.Operator{db}
	usr, err := o.ReadUserByID(userID)
	if err != nil {
		return 0, err
	}

	streamname := usr.Name + "/user/log"

	//Now attempt to go straight for the log stream
	logstream, err := o.ReadStream(streamname)
	if err != nil {
		//We had an error - try to create the stream (the user device is assumed to exist)
		err = o.CreateStream(streamname, UserlogSchema)
		if err != nil {
			return 0, err
		}

		//Now try to read the
		logstream, err = o.ReadStream(streamname)
	}
	return logstream.StreamId, err
}
**/
