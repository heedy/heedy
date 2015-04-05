// Package users provides an API for managing user information.
package users

import (
	"errors"
	"log"
	)

type PermissionLevel string


const (
	NOBODY = PermissionLevel("nobody") // Highest permission level, no device can modify must do it straight in DB
	ROOT = PermissionLevel("root")  // Highest interface permission level or above
	USER = PermissionLevel("user") // Users can modify their own stuff or above
	DEVICE = PermissionLevel("device") // the owning device of a given stream or above
	ENABLED = PermissionLevel("enabled") // any enabled device
	ANYBODY = PermissionLevel("anybody") // lowest permission level, any user logged in or not
)


func strToPermissionLevel(s string) (PermissionLevel, error) {
	pl := PermissionLevel(s)
	switch pl {
		case NOBODY:
			return NOBODY, nil
		case ROOT:
			return ROOT, nil
		case USER:
			return USER, nil
		case DEVICE:
			return DEVICE, nil
		case ENABLED:
			return ENABLED, nil
		case ANYBODY:
			return ANYBODY, nil
	}

	return ANYBODY, errors.New("Given string is not a valid permission type")
}

// Checks that the given permission is at least what the desired one should be
func PermissionLevelGte(actual, desired PermissionLevel) bool {
	switch desired {
		case NOBODY:
			return false
		case ROOT:
			return actual == ROOT
		case USER:
			return actual == ROOT || actual == USER
		case DEVICE:
			return actual == ROOT || actual == USER || actual == DEVICE
		case ENABLED:
			return actual == ROOT || actual == USER || actual == DEVICE || actual == ENABLED
		case ANYBODY:
			return true
	}

	log.Printf("Error, used invalid permission level actual: %v, desired: %v", actual, desired)
	return false
}


// Meta information about streamdb
// for example, the database version
type StreamdbMeta struct {
	Key string `modifiable:"root"`
	Value string `modifiable:"root"`
}

// A per-user KV store
type UserKeyValue struct {
	UserId int64
	Key string `modifiable:"root"`
	Value string `modifiable:"user"`
}

// A per-stream KV store
type StreamKeyValue struct {
	StreamId int64
	Key string `modifiable:"root"`
	Value string `modifiable:"device"`
}

// A per-device KV store
type DeviceKeyValue struct {
	DeviceId int64
	Key string `modifiable:"root"`
	Value string `modifiable:"device"`
}

// User is the storage type for rows of the database.
type User struct {
	UserId    int64  `modifiable:"nobody"` // The primary key
	Name  string `modifiable:"root"`   // The public username of the user
	Email string `modifiable:"user"`   // The user's email address

	Password           string `modifiable:"user"` // A hash of the user's password
	PasswordSalt       string `modifiable:"user"` // The password salt to be attached to the end of the password
	PasswordHashScheme string `modifiable:"user"` // A string representing the hashing scheme used

	Admin        bool   `modifiable:"root"` // True/False if this is an administrator

	UploadLimit_Items int `modifiable:"root"` // upload limit in items/day
	ProcessingLimit_S int `modifiable:"root"` // processing limit in seconds/day
	StorageLimit_Gb   int `modifiable:"root"` // storage limit in GB
}

// Sets a new password for an account
func (u *User) SetNewPassword(newPass string) {
	u.Password = calcHash(newPass, u.PasswordSalt, u.PasswordHashScheme)
}

// Checks if the device is enabled and a superdevice
func (u *User) IsAdmin() bool {
	return u.Admin
}


func (u *User) ValidatePassword(password string) bool {
	return calcHash(password, u.PasswordSalt, u.PasswordHashScheme) == u.Password
}


// Devices are general purposed external and internal data users,
//
type Device struct {
	DeviceId          int64  `modifiable:"nobody"` // The primary key of this device
	Name        string `modifiable:"nobody"` // The registered name of this device, should be universally unique like "Devicename_serialnum"
	Nickname   string `modifiable:"user"`   // The human readable name of this device
	UserId     int64  `modifiable:"root"`   // the user that owns this device
	ApiKey      string `modifiable:"user"`   // A uuid used as an api key to verify against
	Enabled     bool   `modifiable:"user"`   // Whether or not this device can do reading and writing
	IsAdmin bool   `modifiable:"root"`   // Whether or not this is a "superdevice" which has access to the whole API
	CanWrite         bool `modifiable:"user"` // Can this device write to streams? (inactive right now)
	CanWriteAnywhere bool `modifiable:"user"` // Can this device write to others streams? (inactive right now)
	CanActAsUser        bool `modifiable:"user"` // Can this device operate as a user? (inactive right now)
	IsVisible bool `modifiable:"root"`
	UserEditable bool `modifiable:"root"`
}


func (d *Device) RelationToUser(user *User, err error) (PermissionLevel, error)  {
	// guards
	if user == nil {
		return ANYBODY, errors.New("Nil user")
	}

	if err != nil {
		return ANYBODY, err
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT, err
	}

	if d.UserId == user.UserId {
		if d.CanActAsUser {
			return USER, err
		}

		return DEVICE, err
	}

	return ANYBODY, nil
}

/**
func (d *Device) RelationToStream(stream *Stream, err error) (PermissionLevel, error)  {
	// guards
	if stream == nil {
		return ANYBODY, errors.New("Nil stream")
	}

	if err != nil {
		return ANYBODY, err
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT, err
	}

	if d.UserId == user.UserId {

		streamUser, err := stream.ReadUser()
		if err != nil {
			return err
		}

		userproxy := d.CanWriteAnywhere && d.UserId == streamUser.UserId
		if d.CanActAsUser || userproxy {
			return USER, err
		}

		return DEVICE, err
	}

	return ANYBODY, err
}


**/

// Check if the device is enabled
func (d *Device) IsActive() bool {
	return d.Enabled
}


func (d *Device) WriteAllowed() bool {
	return d.CanWrite
}

func (d *Device) WriteAnywhereAllowed() bool {
	return d.CanWriteAnywhere
}

func (d *Device) IsOwnedBy(user *User) bool {
	return d.UserId == user.UserId
}
