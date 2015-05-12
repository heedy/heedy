package streamdb

import (
	"errors"
	"streamdb/users"
	"strings"
)

var (
	//ErrPermissions is thrown when an operator tries to do stuff it is not allowed to do
	ErrPermissions = errors.New("Access Denied")
)

//AuthOperator is the database proxy for a particular device.
//TODO: Operator does not auto-expire after time period
type AuthOperator struct {
	Db *Database //Db is the underlying database

	operatorPath string

	//This ensures that name-changes cannot be exploited
	devID int64 //the id of the device
}

//Name is the path to the device underlying the operator
func (o *AuthOperator) Name() string {
	return o.operatorPath
}

//Reload the device from database
func (o *AuthOperator) Reload() error {
	o.Db.userCache.RemoveID(o.devID)
	return nil
}

//Database returns the underlying database
func (o *AuthOperator) Database() *Database {
	return o.Db
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

//SetAdmin does exactly what it claims. It works on both users and devices
func (o *AuthOperator) SetAdmin(path string, isadmin bool) error {
	switch strings.Count(path, "/") {
	default:
		return ErrBadPath
	case 0:
		u, err := o.ReadUser(path)
		if err != nil {
			return err
		}
		u.Admin = isadmin
		return o.UpdateUser(u)
	case 1:
		dev, err := o.ReadDevice(path)
		if err != nil {
			return err
		}
		dev.IsAdmin = isadmin
		return o.UpdateDevice(dev)
	}
}
