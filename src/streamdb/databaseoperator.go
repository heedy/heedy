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

	//ErrNotChangeable is thrown when changing a field that can't be changed
	ErrNotChangeable = errors.New("The given fields are not modifiable.")
)

//Database just returns self
func (o *Database) Database() *Database {
	return o
}

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
	_, err := o.ReadUser(username)
	if err != nil {
		return err //Workaround for issue #81
	}
	return o.Userdb.DeleteUserByName(username)
}

//UpdateUser performs the given modifications
func (o *Database) UpdateUser(user *users.User, modifieduser users.User) error {
	if modifieduser.RevertUneditableFields(*user, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	return o.Userdb.UpdateUser(&modifieduser)
}

//SetAdmin does exactly what it claims
func (o *Database) SetAdmin(path string, isadmin bool) error {

	//TODO: Make this work with devices
	u, err := o.ReadUser(path)
	if err != nil {
		return err
	}

	modu := *u //Make a copy of the user
	modu.Admin = isadmin

	return o.UpdateUser(u, modu)

}

//ChangeUserPassword changes the password for the given user
func (o *Database) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	modu := *u
	modu.Password, modu.PasswordSalt, modu.PasswordHashScheme = users.UpgradePassword(newpass)

	return o.UpdateUser(u, modu)
}
