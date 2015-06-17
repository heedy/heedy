package authoperator

import (
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"
	"errors"
)

var (
	//ErrPermissions is thrown when an operator tries to do stuff it is not allowed to do
	ErrPermissions = errors.New("Access Denied")
)

//AuthOperator is the database proxy for a particular device.
//TODO: Operator does not auto-expire after time period
type AuthOperator struct {
	Db operator.BaseOperatorInterface //The operator which is used to interact with the database

	operatorPath string //The operator path is the string name of the operator
	devID        int64  //the id of the device - operatorPath is not enough, since name changes can happen in other threads
}

//NewAuthOperator creates a new authenticated operator,
func NewAuthOperator(db operator.BaseOperatorInterface, deviceID int64) (operator.Operator, error) {
	dev, err := db.ReadDeviceByID(deviceID)
	if err != nil {
		return operator.Operator{}, err
	}
	usr, err := db.ReadUserByID(dev.UserId)
	if err != nil {
		return operator.Operator{}, err
	}
	return operator.Operator{&AuthOperator{db, usr.Name + "/" + dev.Name, dev.DeviceId}}, nil
}

//Name is the path to the device underlying the operator
func (o *AuthOperator) Name() string {
	return o.operatorPath
}

//User returns the current user
func (o *AuthOperator) User() (usr *users.User, err error) {
	dev, err := o.Db.ReadDeviceByID(o.devID)
	if err != nil {
		return nil, err
	}
	return o.Db.ReadUserByID(dev.UserId)
}

//Device returns the current device
func (o *AuthOperator) Device() (*users.Device, error) {
	return o.Db.ReadDeviceByID(o.devID)
}

//Permissions returns whether the operator has permissions given by the string
func (o *AuthOperator) Permissions(perm users.PermissionLevel) bool {
	dev, err := o.Device()
	if err != nil {
		return false
	}
	return dev.GeneralPermissions().Gte(perm)
}
