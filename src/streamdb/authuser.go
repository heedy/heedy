package streamdb

import (
	"errors"
	"streamdb/users"
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

//CreateUser makes a new user
func (o *AuthOperator) CreateUser(username, email, password string) error {
	if !o.Permissions(users.ROOT) {
		return ErrPermissions
	}
	return o.Db.CreateUser(username, email, password)
}

//ReadAllUsers reads all the users
func (o *AuthOperator) ReadAllUsers() ([]users.User, error) {
	if o.Permissions(users.ROOT) {
		return o.Db.ReadAllUsers()
	}
	//If not admin, then we only know about our own device
	u, err := o.User()
	if err != nil {
		return []users.User{}, err
	}
	return []users.User{*u}, err
}

//ReadUser reads a user - or rather reads any user that this device has permissions to read
func (o *AuthOperator) ReadUser(username string) (*users.User, error) {
	if o.usrName == username {
		return o.User()
	}
	if o.Permissions(users.ROOT) {
		return o.Db.ReadUser(username)
	}
	return nil, ErrPermissions
}

//ReadUserByEmail reads a user - or rather reads any user that this device has permissions to read
func (o *AuthOperator) ReadUserByEmail(email string) (*users.User, error) {
	u, err := o.User()
	if err != nil {
		return nil, err
	}
	if u.Email == email {
		return u, nil
	}
	if o.Permissions(users.ROOT) {
		return o.Db.ReadUserByEmail(email)
	}
	return nil, ErrPermissions
}

//DeleteUser deletes the given user - only admin can delete
func (o *AuthOperator) DeleteUser(username string) error {
	if !o.Permissions(users.ROOT) {
		return ErrPermissions
	}
	return o.Db.DeleteUser(username)
}

//UpdateUser performs the given modifications
func (o *AuthOperator) UpdateUser(username string, modifieduser *users.User) error {
	user, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	dev, err := o.Device()
	if err != nil {
		return err
	}
	//See if the bastards tried to change a field they have no fucking business editing :-P
	if modifieduser.RevertUneditableFields(*user, dev.RelationToUser(user)) > 0 {
		return ErrPermissions
	}
	//Thankfully, ReadUser put this user right on top of the cache, so it should still be there
	o.Db.UpdateUser(username, modifieduser)
	return err
}

//SetAdmin does exactly what it claims
func (o *AuthOperator) SetAdmin(path string, isadmin bool) error {

	//TODO: Make this work with devices
	u, err := o.ReadUser(path)
	if err != nil {
		return err
	}
	u.Admin = isadmin
	return o.UpdateUser(u.Name, u)

}

//ChangeUserPassword changes the password for the given user
func (o *AuthOperator) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	u.SetNewPassword(newpass)
	return o.UpdateUser(username, u)
}
