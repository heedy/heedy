/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"database/sql"
	"encoding/base64"
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

// UserMaker is the structure used to create users
type UserMaker struct {
	User

	// The devices to create for the user
	Devices map[string]*DeviceMaker `json: "devices"`

	// The streams to create for the user device
	Streams map[string]*StreamMaker `json:"streams"`

	Userlimit int64 `json:"-"`
}

func (um *UserMaker) Validate(deviceLimit int, streamLimit int) error {
	if um == nil {
		return errors.New("null user creation struct")
	}
	// Check the underlying user for validity, after filling in some gibberish for the password hash + salt
	um.PasswordHashScheme = "NULL"
	um.PasswordSalt = "NULL"
	err := um.ValidityCheck()
	if err != nil {
		return err
	}

	if deviceLimit > 0 && len(um.Devices) > deviceLimit {
		return errors.New("Exceeded device limit")
	}
	if streamLimit > 0 && len(um.Streams) > streamLimit {
		return errors.New("Exceeded stream limit for user")
	}
	for d := range um.Devices {
		if d == "user" {
			return errors.New("user device is created by default")
		}
		if d == "meta" {
			return errors.New("meta device is created by default")
		}
		um.Devices[d].Name = d
		if err := um.Devices[d].Validate(streamLimit); err != nil {
			return err
		}
	}

	for s := range um.Streams {
		um.Streams[s].Name = s
		err = um.Streams[s].Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) String() string {
	return fmt.Sprintf("[users.User | Id: %v, Name: %v, Email: %v, Nick: %v, Passwd: %v|%v|%v ]",
		u.UserID, u.Name, u.Email, u.Nickname, u.Password, u.PasswordSalt, u.PasswordHashScheme)
}

// Ensures that the icon is a base64 encoded image
func validateIcon(icon string) error {
	if icon == "" {
		return nil
	}
	_, err := base64.URLEncoding.DecodeString(icon)
	return err
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

	if u.PasswordSalt == "" || u.PasswordHashScheme == "" || u.Password == "" {
		return ErrInvalidPassword
	}

	err = validateIcon(u.Icon)
	if err != nil {
		return err
	}

	if u.Role == "" {
		return errors.New("Role not set for user")
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
// It is assumed that the usermaker was validated (usermaker.Validate() was called)
func (userdb *SqlUserDatabase) CreateUser(um *UserMaker) error {
	if um.Userlimit > 0 {
		// TODO: This check should be done within the SQL transaction to avoid timing attacks
		num, err := userdb.CountUsers()
		if err != nil {
			return err
		}
		if num >= um.Userlimit {
			return ErrMaxUsers
		}
	}

	dbpass, salt, hashtype, err := HashPassword(um.Password)
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
		Public,
		Description,
		Icon,
		Nickname) VALUES (?,?,?,?,?,?,?,?,?,?);`,
		um.Name,
		um.Email,
		dbpass,
		salt,
		hashtype,
		um.Role,
		um.Public,
		um.Description,
		um.Icon,
		um.Nickname)

	if err != nil && strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint ") {
		return errors.New("User with this email or username already exists")
	}
	if err != nil {
		return err
	}

	if len(um.Streams) > 0 || len(um.Devices) > 0 {
		// TODO: Multiple-inserts should be all done in a transaction so that inserting is not super slow
		u, err := userdb.ReadUserByName(um.Name)
		if err != nil {
			return err
		}

		// User creation succeeded - now make all the streams for the user device
		if len(um.Streams) > 0 {

			d, err := userdb.ReadUserOperatingDevice(u)
			if err != nil {
				return err
			}
			for s := range um.Streams {
				um.Streams[s].Name = s
				um.Streams[s].DeviceID = d.DeviceID
				if err = userdb.CreateStream(um.Streams[s]); err != nil {
					userdb.DeleteUser(u.UserID)
					return err
				}
			}
		}

		// We create all of the requested devices for the user
		for d := range um.Devices {
			um.Devices[d].Name = d
			um.Devices[d].UserID = u.UserID
			if err = userdb.CreateDevice(um.Devices[d]); err != nil {
				userdb.DeleteUser(u.UserID)
				return err
			}
		}
	}

	return nil
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
