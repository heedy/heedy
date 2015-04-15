// Package users provides an API for managing user information.
package users

import (
	"errors"
	"reflect"
	)

type PermissionLevel uint

type DatabaseType struct {

}

const (
	NOBODY = PermissionLevel(6) // Highest permission level, no device can modify must do it straight in DB
	ROOT = PermissionLevel(5)  // Highest interface permission level or above
	USER = PermissionLevel(4) // Users can modify their own stuff or above
	DEVICE = PermissionLevel(3) // the owning device of a given stream or above
	FAMILY = PermissionLevel(2) // a device that is a sibbling of the given device or an aunt to a stream
	ENABLED = PermissionLevel(1) // any enabled device
	ANYBODY = PermissionLevel(0) // lowest permission level, any user logged in or not
)


func strToPermissionLevel(s string) (PermissionLevel, error) {
	switch s {
		case "nobody":
			return NOBODY, nil
		case "root":
			return ROOT, nil
		case "user":
			return USER, nil
		case "device":
			return DEVICE, nil
		case "family":
			return FAMILY, nil
		case "enabled":
			return ENABLED, nil
		case "anybody":
			return ANYBODY, nil
	}

	return ANYBODY, errors.New("Given string is not a valid permission type")
}

// Checks that the given permission is at least what the desired one should be
func (actual PermissionLevel)Gte(desired PermissionLevel) bool {
	return uint(actual) >= uint(desired)
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
	DatabaseType
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

func (d *Device) GeneralPermissions() (PermissionLevel) {
	if ! d.Enabled {
		return ANYBODY
	}

	if d.IsAdmin {
		return ROOT
	}

	return ENABLED
}


func (d *Device) RelationToUser(user *User) (PermissionLevel)  {
	// guards
	if user == nil || ! d.Enabled {
		return ANYBODY
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT
	}

	if d.UserId == user.UserId {
		if d.CanActAsUser {
			return USER
		}

		return DEVICE
	}

	return ANYBODY
}



func (d *Device) RelationToDevice(device *Device) (PermissionLevel)  {
	// guards
	if device == nil || ! d.Enabled {
		return ANYBODY
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT
	}

	if d.UserId == device.UserId {
		if d.CanActAsUser {
			return USER
		}

		if d.DeviceId == device.DeviceId {
			return DEVICE
		}

		return FAMILY
	}


	if d.Enabled {
		return ENABLED
	}

	return ANYBODY
}


func (d *Device) RelationToStream(stream *Stream, streamParent *Device) (PermissionLevel)  {
	// guards
	if stream == nil || streamParent == nil || ! d.Enabled {
		return ANYBODY
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT
	}

	if d.CanActAsUser && d.UserId == streamParent.UserId {
		return USER
	}

	if d.DeviceId == stream.DeviceId {
		return DEVICE
	}

	if d.UserId == streamParent.UserId {
		return FAMILY
	}

	return ENABLED
}

func (d *Device) RevertUneditableFields(originalValue Device, p PermissionLevel) {
	// TODO Introspection
	revertUneditableFields(d, originalValue, p)
}

func revertUneditableFields(toChange interface{}, originalValue interface{}, p PermissionLevel) {
	originalValueReflect := reflect.ValueOf(originalValue).Elem()
	toChangeReflect  := reflect.ValueOf(toChange).Elem()

	for i := 0; i < originalValueReflect.NumField(); i++ {
		// Grab the fields for reflection
		originalValueField := originalValueReflect.Field(i)
		toChangeValueField := toChangeReflect.Field(i)

		// Check what kind of modifiable permission we need to edit
		modifiable := originalValueField.Type().Field(i).Tag.Get("modifiable")

		// By default, we don't allow modification
		if modifiable == "" {
			modifiable = "nobody"
		}

		// If we don't have enough permissions, reset the field from original
		requiredPermissionsForField, _ := strToPermissionLevel(modifiable)
		if ! p.Gte(requiredPermissionsForField) && originalValueField.IsValid() && originalValueField.CanSet() {
			toChangeValueField.Set(originalValueField)
		}
	}

	// and bob's your uncle!
}

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
