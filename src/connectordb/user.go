package connectordb

import (
	pconfig "config/permissions"
	"connectordb/users"
	"errors"
	"fmt"
)

// CountUsers returns the total number of users of the entire database
func (db *Database) CountUsers() (int64, error) {
	return db.Userdb.CountUsers()
}

// ReadAllUsers returns all users of the database. This one will be a pretty expensive operation,
// since there is a virtually unlimited number of users possible, so this should have some type of
// search possibility in the future
func (db *Database) ReadAllUsers() ([]*users.User, error) {
	return db.Userdb.ReadAllUsers()
}

// CreateUser creates a user with the given information. It checks some basic validity
// before creating, and ensures that roles exist/max user amounts are upheld
func (db *Database) CreateUser(u *users.UserMaker) error {
	perm := pconfig.Get()

	if !perm.IsAllowedUsername(u.Name) {
		return fmt.Errorf("Username '%s' not allowed", u.Name)
	}

	if !perm.IsAllowedEmail(u.Email) {
		return fmt.Errorf("Email '%s' not allowed", u.Email)
	}

	// Make sure that the given role exists
	r, ok := perm.UserRoles[u.Role]
	if !ok {
		return fmt.Errorf("The given role '%s' does not exist", u.Role)
	}

	// Make sure that users with this role are allowed to be private if private is set
	if !u.Public && !r.CanBePrivate {
		return fmt.Errorf("Users with role '%s' can't be private.", u.Role)
	}

	// Perform user-level validation before creating the user
	if err := u.Validate(int(r.MaxDevices), int(r.MaxStreams)); err != nil {
		return err
	}
	// Set the user limit
	u.Userlimit = perm.MaxUsers
	return db.Userdb.CreateUser(u)

}

// ReadUserByID reads the user object by ID given
func (db *Database) ReadUserByID(userID int64) (*users.User, error) {
	return db.Userdb.ReadUserById(userID)
}

// ReadUser reads the user object by user name
func (db *Database) ReadUser(username string) (*users.User, error) {
	return db.Userdb.ReadUserByName(username)
}

// UpdateUserByID updates the user with the given UserID with the given data map
func (db *Database) UpdateUserByID(userID int64, update map[string]interface{}) error {
	u, err := db.ReadUserByID(userID)
	if err != nil {
		return err
	}

	oldname := u.Name
	_, haspassword := update["password"]

	err = WriteObjectFromMap(u, update)
	if err != nil {
		return err
	}

	// Now ensure that the updated user is valid
	if u.Name != oldname {
		return errors.New("ConnectorDB does not support modification of user names")
	}

	perm := pconfig.Get()

	if !perm.IsAllowedEmail(u.Email) {
		return fmt.Errorf("Email '%s' not allowed", u.Email)
	}

	r, ok := perm.UserRoles[u.Role]
	if !ok {
		return fmt.Errorf("The given role '%s' does not exist", u.Role)
	}

	if !u.Public && !r.CanBePrivate {
		return fmt.Errorf("User can't be private.")
	}

	if haspassword {
		u.SetNewPassword(u.Password)
	}

	return db.Userdb.UpdateUser(u)
}

// DeleteUserByID removes the user with the given UserID. It propagates deletion to add devices
// and streams that the user owns
func (db *Database) DeleteUserByID(userID int64) error {

	// First take all the devices that the user owns, and delete them
	devices, err := db.ReadAllDevicesByUserID(userID)
	if err != nil {
		return err
	}

	for i := range devices {
		err := db.DeleteDeviceByID(devices[i].DeviceID)
		if err != nil {
			return err
		}
	}

	// Lastly, delete the user
	return db.Userdb.DeleteUser(userID)
}
