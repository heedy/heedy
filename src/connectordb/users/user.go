/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

var (
	ErrInvalidUsername = errors.New("Invalid Username, usernames may not contain / \\ ? or spaces")
	ErrInvalidEmail    = errors.New("Invalid Email Address")
	ErrEmailExists     = errors.New("A user already exists with this email")
	ErrUsernameExists  = errors.New("A user already exists with this username")
	ErrDisallowedEmail = errors.New("The email domain you specified is not valid")
	ErrMaxUsers        = errors.New("Maximum user limit was reached")
)

// User is the storage type for rows of the database.
type User struct {
	UserID int64 `json:"-"` // The primary key

	Name        string `json:"name"`        // The public username of the user
	Nickname    string `json:"nickname"`    // The nickname of the user
	Email       string `json:"email"`       // The user's email address
	Description string `json:"description"` // A public description
	Icon        string `json:"icon"`        // A public icon in a data URI format, should be smallish 100x100?

	Role   string `json:"role,omitempty"` // The user type (permissions level)
	Public bool   `json:"public"`         // Whether the user is public or not

	Password           string `json:"password,omitempty"` // A hash of the user's password - it is never actually returned - the json params are used internally
	PasswordSalt       string `json:"-"`                  // The password salt to be attached to the end of the password
	PasswordHashScheme string `json:"-"`                  // A string representing the hashing scheme used

}

func (u *User) String() string {
	return fmt.Sprintf("[users.User | Id: %v, Name: %v, Email: %v, Nick: %v, Passwd: %v|%v|%v ]",
		u.UserID, u.Name, u.Email, u.Nickname, u.Password, u.PasswordSalt, u.PasswordHashScheme)
}

// ValidityCheck checks if the fields are valid, e.g. we're not trying to change the name to blank.
func (u *User) ValidityCheck() error {
	if !IsValidName(u.Name) {
		return ErrInvalidUsername
	}

	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return ErrInvalidEmail
	}

	if u.PasswordSalt == "" || u.PasswordHashScheme == "" {
		return ErrInvalidPassword
	}

	// NOTE: we DO NOT check for allowed email domains here, a user can change
	// their preferred email address once they're in the system

	return nil
}

// SetNewPassword sets a new password for an account
func (u *User) SetNewPassword(newPass string) error {
	hash, salt, scheme, err := HashPassword(newPass)
	if err != nil {
		return err
	}

	u.PasswordHashScheme = scheme
	u.PasswordSalt = salt
	u.Password = hash
	return nil
}

// ValidatePassword returns true if password matches
func (u *User) ValidatePassword(password string) bool {
	return CheckPassword(password, u.Password, u.PasswordSalt, u.PasswordHashScheme) == nil
}

// UpgradePassword upgrades the security of the password, returns True if the user needs to be
// saved again because an upgrade was performed.
func (u *User) UpgradePassword(password string) bool {
	if !UpgradePassword(u.Password, u.PasswordSalt, u.PasswordHashScheme) {
		return false
	}

	hash, salt, scheme, err := HashPassword(password)
	if err != nil {
		// Uh oh... Since creating a hash failed, return false
		return false
	}

	u.PasswordHashScheme = scheme
	u.PasswordSalt = salt
	u.Password = hash

	return true
}

// CreateUser creates a user given the user's credentials.
// If a user already exists with the given credentials, an error is thrown.
func (userdb *SqlUserDatabase) CreateUser(Name, Email, Password, Role string, Public bool, userlimit int64) error {
	/*
		existing, err := userdb.readByNameOrEmail(Name, Email)

		if err == nil {
			// Check for existence of user to provide helpful notices

			switch {
			case existing.Email == Email:
				return ErrEmailExists
			case existing.Name == Name:
				return ErrUsernameExists

			}
		}*/

	switch {
	case !IsValidName(Name):
		return ErrInvalidUsername
	case userlimit > 0:
		// TODO: This check should be done within the SQL transaction to avoid timing attacks
		num, err := userdb.CountUsers()
		if err != nil {
			return err
		}
		if num >= userlimit {
			return ErrMaxUsers
		}
	}

	dbpass, salt, hashtype, err := HashPassword(Password)
	if err != nil {
		return err
	}

	_, err = userdb.Exec(`INSERT INTO Users (
		Name,
		Email,
		Password,
		PasswordSalt,
		PasswordHashScheme,
		Role,
		Public) VALUES (?,?,?,?,?,?,?);`,
		Name,
		Email,
		dbpass,
		salt,
		hashtype,
		Role, Public)

	if err != nil && strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint ") {
		return errors.New("User with this email or username already exists")
	}

	return err
}

/*Login Performs a login function on the user.

Looks for a user by the (username|email)/password pair.
Checks the password, if it's a match, tries to upgrade the password.
Finally, grabs the User device for performing user actions from.

Returns an error along with the user and device if something went wrong

*/
func (userdb *SqlUserDatabase) Login(Username, Password string) (*User, *Device, error) {
	user, err := userdb.readByNameOrEmail(Username, Username)
	if err != nil {
		return nil, nil, ErrInvalidUsername
	}

	if !user.ValidatePassword(Password) {
		return user, nil, ErrInvalidPassword
	}

	if user.UpgradePassword(Password) {
		userdb.UpdateUser(user)
	}

	opdev, err := userdb.ReadUserOperatingDevice(user)

	return user, opdev, err
}

// Reads the operating device for the user (the implicity device the user uses)
func (userdb *SqlUserDatabase) ReadUserOperatingDevice(user *User) (*Device, error) {
	return userdb.ReadDeviceForUserByName(user.UserID, "user")
}

// readByNameOrEmail returns a User instance if a user exists with the given
// email address or username
func (userdb *SqlUserDatabase) readByNameOrEmail(Name, Email string) (*User, error) {
	var exists User

	err := userdb.Get(&exists, "SELECT * FROM Users WHERE upper(Name) = upper(?) OR upper(Email) = upper(?) LIMIT 1;", Name, Email)

	//err := userdb.Get(&exists, "SELECT * FROM Users WHERE Name = ? OR upper(Email) = upper(?) LIMIT 1;", Name, Email)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return &exists, err
}

// ReadUserByName returns a User instance if a user exists with the given
// username.
func (userdb *SqlUserDatabase) ReadUserByName(Name string) (*User, error) {
	var user User

	err := userdb.Get(&user, "SELECT * FROM Users WHERE Name = ? LIMIT 1;", Name)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return &user, err
}

// ReadUserById returns a User instance if a user exists with the given
// id.
func (userdb *SqlUserDatabase) ReadUserById(UserID int64) (*User, error) {
	var user User
	err := userdb.Get(&user, "SELECT * FROM Users WHERE UserID = ? LIMIT 1;", UserID)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return &user, err
}

func (userdb *SqlUserDatabase) ReadAllUsers() ([]*User, error) {
	var users []*User

	err := userdb.Select(&users, "SELECT * FROM Users")

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

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

	_, err := userdb.Exec(`UPDATE users SET
					Name=?,
					Nickname=?,
					Email=?,
					Password=?,
					PasswordSalt=?,
					PasswordHashScheme=?,
					Description=?,
					Icon=?,
					Public=?,
					Role=?
					WHERE UserID = ?`,
		user.Name,
		user.Nickname,
		user.Email,
		user.Password,
		user.PasswordSalt,
		user.PasswordHashScheme,
		user.Description,
		user.Icon,
		user.Public,
		user.Role,
		user.UserID)

	return err
}

// DeleteUser removes a user from the database
func (userdb *SqlUserDatabase) DeleteUser(UserID int64) error {
	result, err := userdb.Exec(`DELETE FROM Users WHERE UserID = ?;`, UserID)
	return getDeleteError(result, err)
}
