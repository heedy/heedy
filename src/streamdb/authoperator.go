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

	usrName string //The user name underlying this device
	devName string //The device name underlying this device
}

//Name is the path to the device underlying the operator
func (o *AuthOperator) Name() string {
	return o.usrName + "/" + o.devName
}

//Reload both user and device
func (o *AuthOperator) Reload() error {
	o.Db.userCache.Remove(o.usrName)
	o.Db.deviceCache.Remove(o.Name())
	return nil
}

//Database returns the underlying database
func (o *AuthOperator) Database() *Database {
	return o.Db
}

//User returns the current user
func (o *AuthOperator) User() (usr *users.User, err error) {
	return o.Db.ReadUser(o.usrName)
}

//Device returns the current device
func (o *AuthOperator) Device() (*users.Device, error) {
	return o.Db.ReadDevice(o.Name())
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
	parray := strings.Split(path, "/")
	if len(parray) > 2 {
		return ErrBadPath
	}
	if len(parray) == 2 { //This is a device
		dev, err := o.ReadDevice(path)
		if err != nil {
			return err
		}
		dev.IsAdmin = isadmin
		return o.UpdateDevice(path, dev)
	}
	//It is a user
	u, err := o.ReadUser(path)
	if err != nil {
		return err
	}
	u.Admin = isadmin
	return o.UpdateUser(u.Name, u)

}
