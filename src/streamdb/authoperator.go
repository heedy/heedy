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

//Loads the device from database
func (o *AuthOperator) reloadDevice() error {
	dev, err := o.Db.Userdb.ReadDeviceById(o.dev.DeviceId)
	if err != nil {
		return err
	}
	o.dev = dev
	return err
}

//Loads the user from database
func (o *AuthOperator) reloadUser() error {
	usr, err := o.Db.Userdb.ReadUserById(o.usr.UserId)
	if err != nil {
		return err
	}
	o.usr = usr
	return err
}

//Reload both user and device
func (o *AuthOperator) Reload() error {
	if err := o.reloadUser(); err != nil {
		return err
	}
	return o.reloadDevice()
}

//Database returns the underlying database
func (o *AuthOperator) Database() *Database {
	return o.Db
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
	if o.usr.Name == username {
		return o.usr, nil
	}
	if o.Permissions(users.ROOT) {
		return o.Db.ReadUser(username)
	}
	return nil, ErrPermissions
}

//ReadUserByEmail reads a user - or rather reads any user that this device has permissions to read
func (o *AuthOperator) ReadUserByEmail(email string) (*users.User, error) {
	if o.usr.Email == email {
		return o.usr, nil
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
func (o *AuthOperator) UpdateUser(user *users.User, modifieduser users.User) error {
	//See if the bastards tried to change a field they have no fucking business editing :-P
	if modifieduser.RevertUneditableFields(*user, o.dev.RelationToUser(user)) > 0 {
		return ErrPermissions
	}
	err := o.Db.Userdb.UpdateUser(&modifieduser)
	if err == nil && user.Name == o.usr.Name {
		//o.usr = &modifieduser //If we are modifying self, then save the changes in self also
		return o.Reload() //Since stuff is modified on triggers, reload both user and device on update
	}
	return err
}

//SetAdmin does exactly what it claims
func (o *AuthOperator) SetAdmin(path string, isadmin bool) error {

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
func (o *AuthOperator) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	modu := *u
	modu.Password, modu.PasswordSalt, modu.PasswordHashScheme = users.UpgradePassword(newpass)

	return o.UpdateUser(u, modu)
}
