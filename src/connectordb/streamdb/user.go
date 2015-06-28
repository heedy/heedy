package streamdb

import (
	"connectordb/streamdb/users"
	"errors"
)

/*
These functions allow Database to conform to the Operator interface
*/

var (
	//ErrNotChangeable is thrown when changing a field that can't be changed
	ErrNotChangeable = errors.New("The given fields are not modifiable.")
)

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

//ReadUserByID reads a user by its ID
func (o *Database) ReadUserByID(userID int64) (*users.User, error) {
	return o.Userdb.ReadUserById(userID)
}

//UpdateUser performs the given modifications
func (o *Database) UpdateUser(modifieduser *users.User) error {
	user, err := o.ReadUserByID(modifieduser.UserId)
	if err != nil {
		return err //Workaround for issue #81
	}
	if modifieduser.RevertUneditableFields(*user, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	return o.Userdb.UpdateUser(modifieduser)
}

//DeleteUserByID deletes a user using its ID
func (o *Database) DeleteUserByID(userID int64) error {
	return o.Userdb.DeleteUser(userID)
}
