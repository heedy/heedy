package streamdb

import (
	"errors"
	"streamdb/users"
)

/*
These functions allow Database to conform to the Operator interface
*/

var (
	//ErrAdmin is thrown when trying to get the user or device of the Admin operator
	ErrAdmin = errors.New("An administrative operator has no user or device")
)

//User returns the current user
func (o *Database) User() (usr *users.User, err error) {
	return nil, ErrAdmin
}

//Device returns the current device
func (o *Database) Device() (*users.Device, error) {
	return nil, ErrAdmin
}

//Permissions returns whether the operator has permissions given by the string
func (o *Database) Permissions(perm users.PermissionLevel) bool {
	return true
}

//The following functions are direct mirrors of Userdb

//CreateUser makes a new user
func (o *Database) CreateUser(username, email, password string) error {
	return o.Userdb.CreateUser(username, email, password)
}

//ReadAllUsers reads all the users
func (o *Database) ReadAllUsers() ([]users.User, error) {
	return o.Userdb.ReadAllUsers()
}

//ReadUser reads a user - or rather reads any user that this device has permissions to read
func (o *Database) ReadUser(username string) (*users.User, error) {
	return o.Userdb.ReadUserByName(username)
}

//ReadUserByEmail reads a user - or rather reads any user that this device has permissions to read
func (o *Database) ReadUserByEmail(email string) (*users.User, error) {
	return o.Userdb.ReadUserByEmail(email)
}

//DeleteUser deletes the given user - only admin can delete
func (o *Database) DeleteUser(username string) error {
	return o.Userdb.DeleteUserByName(username)
}
