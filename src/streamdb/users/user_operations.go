// Package users provides an API for managing user information.
package users

import (
	"errors"
)

var (
	InvalidPasswordError = errors.New("Invalid Password")
	InvalidUsernameError = errors.New("Invalid Username")
)

// CreateUser creates a user given the user's credentials.
// If a user already exists with the given credentials, an error is thrown.
func (userdb *UserDatabase) CreateUser(Name, Email, Password string) error {

	existing, _ := userdb.ReadByNameOrEmail(Name, Email)

	// Check for existance of user to provide helpful notices
	switch {
	case existing.Email == Email:
		return ERR_EMAIL_EXISTS
	case existing.Name == Name:
		return ERR_USERNAME_EXISTS
	}

	dbpass, salt, hashtype := UpgradePassword(Password)

	_, err := userdb.Exec(`INSERT INTO Users (
	    Name,
	    Email,
	    Password,
	    PasswordSalt,
	    PasswordHashScheme) VALUES (?,?,?,?,?);`,
		Name,
		Email,
		dbpass,
		salt,
		hashtype)

	return err
}

/** Performs a login function on the user.

Looks for a user by the (username|email)/password pair.
Checks the password, if it's a match, tries to upgrade the password.
Finally, grabs the User device for performing user actions from.

Returns an error along with the user and device if something went wrong

**/
func (userdb *UserDatabase) Login(Username, Password string) (*User, *Device, error) {
	user, err := userdb.ReadByNameOrEmail(Username, Username)
	if err != nil {
		return nil, nil, InvalidUsernameError
	}

	if !user.ValidatePassword(Password) {
		return user, nil, InvalidPasswordError
	}

	if user.UpgradePassword(Password) {
		userdb.UpdateUser(user)
	}

	opdev, err := userdb.ReadUserOperatingDevice(user)

	return user, opdev, err
}

// Reads the operating device for the user (the implicity device the user uses)
func (userdb *UserDatabase) ReadUserOperatingDevice(user *User) (*Device, error) {
	return userdb.ReadDeviceForUserByName(user.UserId, "user")
}

// ReadUserByEmail returns a User instance if a user exists with the given
// email address.
func (userdb *UserDatabase) ReadByNameOrEmail(Name, Email string) (*User, error) {
	var exists User

	err := userdb.Get(&exists, "SELECT * FROM Users WHERE Name = ? OR Email = ? LIMIT 1;", Name, Email)

	return &exists, err
}

// ReadUserByEmail returns a User instance if a user exists with the given
// email address.
func (userdb *UserDatabase) ReadUserByEmail(Email string) (*User, error) {
	var user User

	err := userdb.Get(&user, "SELECT * FROM Users WHERE Email = ? LIMIT 1;", Email)

	return &user, err
}

// ReadUserByName returns a User instance if a user exists with the given
// username.
func (userdb *UserDatabase) ReadUserByName(Name string) (*User, error) {
	var user User

	err := userdb.Get(&user, "SELECT * FROM Users WHERE Name = ? LIMIT 1;", Name)
	return &user, err
}

// ReadUserById returns a User instance if a user exists with the given
// id.
func (userdb *UserDatabase) ReadUserById(UserId int64) (*User, error) {
	var user User
	err := userdb.Get(&user, "SELECT * FROM Users WHERE UserId = ? LIMIT 1;", UserId)

	return &user, err
}

/**
func (userdb *UserDatabase) ReadUsersForDevice(devId uint64) ([]User, error){
	var users []User
	err := userdb.Select(&users, "SELECT u* FROM Users u, Devices d WHERE d.DeviceId = ? AND u.UserId = d.UserId OR ? = TRUE")
}
**/

func (userdb *UserDatabase) ReadAllUsers() ([]User, error) {
	var users []User

	err := userdb.Select(&users, "SELECT * FROM Users")

	return users, err
}

func (userdb *UserDatabase) ReadStreamOwner(StreamId int64) (*User, error) {
	var user User

	err := userdb.Get(&user, `SELECT u.*
	                              FROM Users u, Streams s, Devices d
	                              WHERE s.StreamId = ?
	                                AND d.DeviceId = s.DeviceId
	                                AND u.UserId = d.UserId
	                              LIMIT 1;`, StreamId)

	return &user, err
}

// UpdateUser updates the user with the given id in the database using the
// information provided in the user struct.
func (userdb *UserDatabase) UpdateUser(user *User) error {
	if user == nil {
		return ERR_INVALID_PTR
	}

	_, err := userdb.Exec(`UPDATE Users SET
	                Name=?, Email=?, Password=?, PasswordSalt=?, PasswordHashScheme=?,
	                Admin=?, UploadLimit_Items=?,
	                ProcessingLimit_S=?, StorageLimit_Gb=? WHERE UserId = ?`,
		user.Name,
		user.Email,
		user.Password,
		user.PasswordSalt,
		user.PasswordHashScheme,
		user.Admin,
		user.UploadLimit_Items,
		user.ProcessingLimit_S,
		user.StorageLimit_Gb,
		user.UserId)

	return err
}

// DeleteUser removes a user from the database
func (userdb *UserDatabase) DeleteUser(UserId int64) error {
	_, err := userdb.Exec(`DELETE FROM Users WHERE UserId = ?;`, UserId)
	return err
}

// DeleteUserByName removes a user from the database by name
func (userdb *UserDatabase) DeleteUserByName(Username string) error {
	_, err := userdb.Exec(`DELETE FROM Users WHERE Name = ?;`, Username)
	return err
}
