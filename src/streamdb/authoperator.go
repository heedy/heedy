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

	dev *users.Device //The device underlying this operator. If it is nil, it is an admin
	usr *users.User   //If the underlying user is queried, it is stored here for future reference. Nil by default
}

//User returns the current user
func (o *AuthOperator) User() (usr *users.User, err error) {
	return o.usr, nil
}

//Device returns the current device
func (o *AuthOperator) Device() (*users.Device, error) {
	return o.dev, nil
}

//Permissions returns whether the operator has permissions given by the string
func (o *AuthOperator) Permissions(perm users.PermissionLevel) bool {
	return o.dev.GeneralPermissions().Gte(perm)
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
	return []users.User{*o.usr}, nil
}

//ReadUser reads a user - or rather reads any user that this device has permissions to read
func (o *AuthOperator) ReadUser(username string) (*users.User, error) {
	u, err := o.User()
	if err != nil {
		return nil, err
	}
	if u.Name == username {
		return u, nil
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
