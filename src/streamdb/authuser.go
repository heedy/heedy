package streamdb

import "streamdb/users"

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
		usr, err := o.User()
		if err != nil {
			return nil, err
		}
		//The username could have changed at this moment
		if usr.Name == username {
			return usr, nil
		}
	}
	if o.Permissions(users.ROOT) {
		return o.Db.ReadUser(username)
	}
	return nil, ErrPermissions
}

//ReadUserByID Note: Reading by Id cannot make use of the cache. it ALWAYS touches the database.
//This is a good way to ensure that the cache or expired names don't mess with things
//Note that to ensure correctness, it does not attempt to read a user without root permissions
func (o *AuthOperator) ReadUserByID(userID int64) (*users.User, error) {
	if o.Permissions(users.ROOT) {
		return o.Db.ReadUserByID(userID)
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

//DeleteUserByID deletes the given user - only admin can delete
func (o *AuthOperator) DeleteUserByID(userID int64) error {
	if !o.Permissions(users.ROOT) {
		return ErrPermissions
	}
	return o.Db.DeleteUserByID(userID)
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

//ChangeUserPassword changes the password for the given user
func (o *AuthOperator) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	u.SetNewPassword(newpass)
	return o.UpdateUser(username, u)
}
