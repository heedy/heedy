package streamdb

import (
	"errors"
	"streamdb/users"
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
	//Check if the user is in the cache
	if u, ok := o.userCache.Get(username); ok {
		usr := u.(users.User)
		return &usr, nil
	}

	usr, err := o.Userdb.ReadUserByName(username)
	if err == nil {
		//put the user into the cache
		o.userCache.Add(usr.Name, *usr)
	}
	return usr, err
}

//ReadUserByEmail reads a user - or rather reads any user that this device has permissions to read
func (o *Database) ReadUserByEmail(email string) (*users.User, error) {
	usr, err := o.Userdb.ReadUserByEmail(email)
	if err == nil {
		//put the user into the cache
		o.userCache.Add(usr.Name, *usr)
	}
	return usr, err
}

//DeleteUser deletes the given user - only admin can delete
func (o *Database) DeleteUser(username string) error {
	_, err := o.ReadUser(username)
	if err != nil {
		return err //Workaround for issue #81
	}
	//DeleteUserDevices is not needed for users, but necessary for timebatchdb and cache cleaning
	err = o.DeleteUserDevices(username)
	if err != nil {
		return err
	}

	//BUG(daniel): The behavior here under malicious attack is undefined. In particular, the cache might
	//be in an inconsistent state if a user happens to create a new device at this moment in another thread, and before
	//userdb deletes the user. This will leave a leftover device in the cache, which could allow
	//a new user with the same name to be created immediately, and it just *might* be possible to do some shenanigans
	//with the leftover device in cache. That is why it is important to clear the cache of existing devices

	o.userCache.Remove(username)

	return o.Userdb.DeleteUserByName(username)
}

//UpdateUser performs the given modifications
func (o *Database) UpdateUser(username string, modifieduser *users.User) error {
	user, err := o.ReadUser(username)
	if err != nil {
		return err //Workaround for issue #81
	}
	if modifieduser.RevertUneditableFields(*user, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	err = o.Userdb.UpdateUser(modifieduser)
	if err == nil {
		//The username was changed - remove the old one from cache
		if username != modifieduser.Name {
			o.userCache.Remove(username)
		}
		o.userCache.Add(modifieduser.Name, *modifieduser)
	}
	return err
}

//ChangeUserPassword changes the password for the given user
func (o *Database) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	u.SetNewPassword(newpass)
	return o.UpdateUser(username, u)
}
