package authoperator

import (
	"connectordb/operator/interfaces"
	"connectordb/users"
	"errors"
)

var (
	//ErrPermissions is thrown when an operator tries to do stuff it is not allowed to do
	ErrPermissions = errors.New("Access Denied")
	ErrBadPath     = errors.New("not a valid path")
)

//AuthOperator is the database proxy for a particular device.
//TODO: Operator does not auto-expire after time period
type AuthOperator struct {
	interfaces.BaseOperator //The operator which is used to interact with the database

	operatorPath string //The operator path is the string name of the operator
	devID        int64  //the id of the device - operatorPath is not enough, since name changes can happen in other threads

	metalogID int64 //The ID of the stream which provides the metalog
}

//Name is the path to the device underlying the operator
func (o *AuthOperator) Name() string {
	return o.operatorPath
}

//User returns the current user
func (o *AuthOperator) User() (usr *users.User, err error) {
	dev, err := o.BaseOperator.ReadDeviceByID(o.devID)
	if err != nil {
		return nil, err
	}
	return o.BaseOperator.ReadUserByID(dev.UserId)
}

//Device returns the current device
func (o *AuthOperator) Device() (*users.Device, error) {
	return o.BaseOperator.ReadDeviceByID(o.devID)
}

//Permissions returns whether the operator has permissions given by the string
func (o *AuthOperator) Permissions(perm users.PermissionLevel) bool {
	dev, err := o.Device()
	if err != nil {
		return false
	}
	return dev.GeneralPermissions().Gte(perm)
}

func (o *AuthOperator) getDevicePath(deviceID int64) (path string, err error) {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return "", err
	}

	usr, err := o.ReadUserByID(dev.UserId)
	if err != nil {
		return "", err
	}
	return usr.Name + "/" + dev.Name, nil
}

func (o *AuthOperator) getStreamPath(streamID int64) (path string, err error) {
	s, err := o.ReadStreamByID(streamID)
	if err != nil {
		return "", err
	}
	devpath, err := o.getDevicePath(s.DeviceId)
	return devpath + "/" + s.Name, err
}

//CountUsers returns the total number of users contatined in the database
func (o *AuthOperator) CountUsers() (uint64, error) {
	if o.Permissions(users.ROOT) {
		return o.BaseOperator.CountUsers()
	}
	return 0, ErrPermissions
}

//CountDevices returns the total number of devices contatined in the database
func (o *AuthOperator) CountDevices() (uint64, error) {
	if o.Permissions(users.ROOT) {
		return o.BaseOperator.CountDevices()
	}
	return 0, ErrPermissions
}

//CountStreams returns the total number of streams contatined in the database
func (o *AuthOperator) CountStreams() (uint64, error) {
	if o.Permissions(users.ROOT) {
		return o.BaseOperator.CountStreams()
	}
	return 0, ErrPermissions
}
