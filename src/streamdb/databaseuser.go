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
	if u, ok := o.userCache.GetByName(username); ok {
		usr := u.(users.User)
		return &usr, nil
	}

	usr, err := o.Userdb.ReadUserByName(username)
	if err == nil {
		//put the user into the cache
		o.userCache.Set(usr.Name, usr.UserId, *usr)
	}
	return usr, err
}

//ReadUserByID reads a user by its ID
func (o *Database) ReadUserByID(userID int64) (*users.User, error) {
	//Check if the user is in the cache
	if u, _, ok := o.userCache.GetByID(userID); ok {
		usr := u.(users.User)
		return &usr, nil
	}

	usr, err := o.Userdb.ReadUserById(userID)
	if err == nil {
		//put the user into the cache
		o.userCache.Set(usr.Name, usr.UserId, *usr)
	}
	return usr, err
}

//ReadUserByEmail reads a user - or rather reads any user that this device has permissions to read
func (o *Database) ReadUserByEmail(email string) (*users.User, error) {
	usr, err := o.Userdb.ReadUserByEmail(email)
	if err == nil {
		//put the user into the cache
		o.userCache.Set(usr.Name, usr.UserId, *usr)
	}
	return usr, err
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

	err = o.Userdb.UpdateUser(modifieduser)
	if err == nil {
		o.userCache.Set(modifieduser.Name, modifieduser.UserId, *modifieduser)

		//Modifications to user can modify properties of user device, so
		//clear the device from cache
		dev, err := o.ReadDevice(modifieduser.Name + "/user")
		if err == nil {
			o.deviceCache.RemoveID(dev.DeviceId)
		}
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
	return o.UpdateUser(u)
}

//DeleteUser deletes the given user - only admin can delete
func (o *Database) DeleteUser(username string) error {
	_, err := o.ReadUser(username)
	if err != nil {
		return err //Workaround for issue #81
	}

	//We want the user removed from user cache after it is deleted from UserDB,
	//so that no process can reinsert in while it is deleting
	defer o.userCache.RemoveName(username)

	//This is inefficient but absolutely necessary for not allowing logins from nonexisting devices
	defer o.deviceCache.UnlinkNamePrefix(username + "/")
	return o.Userdb.DeleteUserByName(username)
}

//DeleteUserByID deletes a user using its ID
func (o *Database) DeleteUserByID(userID int64) error {
	usr, err := o.ReadUserByID(userID)
	if err != nil {
		return err //Workaround for issue #81
	}
	err = o.Userdb.DeleteUser(userID)
	if err == nil {
		err = o.tdb.DeletePrefix(getTimebatchUserName(usr.UserId) + "/")
	}
	o.userCache.RemoveID(userID)
	o.deviceCache.UnlinkNamePrefix(usr.Name + "/")
	return err
}
