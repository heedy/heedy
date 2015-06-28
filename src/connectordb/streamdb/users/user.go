// Package users provides an API for managing user information.
package users

import (
	"errors"
	"reflect"
)

var (
	InvalidPasswordError = errors.New("Invalid Password")
	InvalidUsernameError = errors.New("Invalid Username, usernames may not contain / \\ ? or spaces")
	InvalidEmailError    = errors.New("Invalid Email Address")
	EmailExistsError     = errors.New("A user already exists with this email")
	UsernameExistsError  = errors.New("A user already exists with this username")
)

// User is the storage type for rows of the database.
type User struct {
	UserId int64  `modifiable:"nobody" json:"-"`   // The primary key
	Name   string `modifiable:"root" json:"name"`  // The public username of the user
	Email  string `modifiable:"user" json:"email"` // The user's email address

	Password           string `modifiable:"user" json:"password,omitempty"` // A hash of the user's password
	PasswordSalt       string `modifiable:"user" json:"-"`                  // The password salt to be attached to the end of the password
	PasswordHashScheme string `modifiable:"user" json:"-"`                  // A string representing the hashing scheme used

	Admin bool `modifiable:"root" json:"admin,omitempty"` // True/False if this is an administrator

	//Since we temporarily don't use limits, I have disabled cluttering results with them on json output
	UploadLimit_Items int `modifiable:"root" json:"-"` // upload limit in items/day
	ProcessingLimit_S int `modifiable:"root" json:"-"` // processing limit in seconds/day
	StorageLimit_Gb   int `modifiable:"root" json:"-"` // storage limit in GB
}

// Checks if the fields are valid, e.g. we're not trying to change the name to blank.
func (u *User) ValidityCheck() error {
	if !IsValidName(u.Name) {
		return InvalidUsernameError
	}

	if u.Email == "" {
		return InvalidEmailError
	}

	if u.PasswordSalt == "" || u.PasswordHashScheme == "" {
		return InvalidPasswordError
	}

	return nil
}

func (d *User) RevertUneditableFields(originalValue User, p PermissionLevel) int {
	return revertUneditableFields(reflect.ValueOf(d), reflect.ValueOf(originalValue), p)
}

// Sets a new password for an account
func (u *User) SetNewPassword(newPass string) {
	hash, salt, scheme := UpgradePassword(newPass)

	u.PasswordHashScheme = scheme
	u.PasswordSalt = salt
	u.Password = hash
}

// Checks if the device is enabled and a superdevice
func (u *User) IsAdmin() bool {
	return u.Admin
}

func (u *User) ValidatePassword(password string) bool {
	return calcHash(password, u.PasswordSalt, u.PasswordHashScheme) == u.Password
}

// Upgrades the security of the password, returns True if the user needs to be
// saved again because an upgrade was performed.
func (u *User) UpgradePassword(password string) bool {
	hash, salt, scheme := UpgradePassword(password)

	if u.PasswordHashScheme == scheme {
		return false
	}

	u.PasswordHashScheme = scheme
	u.PasswordSalt = salt
	u.Password = hash

	return true
}

// CreateUser creates a user given the user's credentials.
// If a user already exists with the given credentials, an error is thrown.
func (userdb *SqlUserDatabase) CreateUser(Name, Email, Password string) error {

	existing, _ := userdb.readByNameOrEmail(Name, Email)

	// Check for existance of user to provide helpful notices
	switch {
	case existing.Email == Email:
		return EmailExistsError
	case existing.Name == Name:
		return UsernameExistsError
	case !IsValidName(Name):
		return InvalidUsernameError
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
func (userdb *SqlUserDatabase) Login(Username, Password string) (*User, *Device, error) {
	user, err := userdb.readByNameOrEmail(Username, Username)
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
func (userdb *SqlUserDatabase) ReadUserOperatingDevice(user *User) (*Device, error) {
	return userdb.ReadDeviceForUserByName(user.UserId, "user")
}

// readByNameOrEmail returns a User instance if a user exists with the given
// email address or username
func (userdb *SqlUserDatabase) readByNameOrEmail(Name, Email string) (*User, error) {
	var exists User

	err := userdb.Get(&exists, "SELECT * FROM Users WHERE upper(Name) = upper(?) OR upper(Email) = upper(?) LIMIT 1;", Name, Email)

	//err := userdb.Get(&exists, "SELECT * FROM Users WHERE Name = ? OR upper(Email) = upper(?) LIMIT 1;", Name, Email)

	return &exists, err
}

// ReadUserByName returns a User instance if a user exists with the given
// username.
func (userdb *SqlUserDatabase) ReadUserByName(Name string) (*User, error) {
	var user User
	//err := userdb.Get(&user, "SELECT * FROM Users WHERE upper(Name) = upper(?) LIMIT 1;", Name)

	err := userdb.Get(&user, "SELECT * FROM Users WHERE Name = ? LIMIT 1;", Name)

	return &user, err
}

// ReadUserById returns a User instance if a user exists with the given
// id.
func (userdb *SqlUserDatabase) ReadUserById(UserId int64) (*User, error) {
	var user User
	err := userdb.Get(&user, "SELECT * FROM Users WHERE UserId = ? LIMIT 1;", UserId)

	return &user, err
}

/**
func (userdb *SqlUserDatabase) ReadUsersForDevice(devId uint64) ([]User, error){
	var users []User
	err := userdb.Select(&users, "SELECT u* FROM Users u, Devices d WHERE d.DeviceId = ? AND u.UserId = d.UserId OR ? = TRUE")
}
**/

func (userdb *SqlUserDatabase) ReadAllUsers() ([]User, error) {
	var users []User

	err := userdb.Select(&users, "SELECT * FROM Users")

	return users, err
}

// UpdateUser updates the user with the given id in the database using the
// information provided in the user struct.
func (userdb *SqlUserDatabase) UpdateUser(user *User) error {
	if user == nil {
		return InvalidPointerError
	}

	if err := user.ValidityCheck(); err != nil {
		return err
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
func (userdb *SqlUserDatabase) DeleteUser(UserId int64) error {
	result, err := userdb.Exec(`DELETE FROM Users WHERE UserId = ?;`, UserId)
	return getDeleteError(result, err)
}
