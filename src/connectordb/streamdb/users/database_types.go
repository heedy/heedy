// Package users provides an API for managing user information.
package users

import (
	"errors"
	"reflect"
	//"fmt"
)

// A PermissionLevel within the system. These determine which devices can
// edit/read all the data in the database.
type PermissionLevel uint

const (
	// NOBODY is the highest permission level, nobody has it; if something needs
	// to be changed with a NOBODY permission, it must be done using SQL. These
	// operations are to keep dangerous things from happening, like modifying
	// primary keys
	NOBODY = PermissionLevel(6)
	// ROOT is the highest permission level given to admin devices/users
	ROOT = PermissionLevel(5)
	// USER is the permission level given to devices that can modify user data
	// and act on a user's behalf
	USER = PermissionLevel(4)
	// DEVICE the device that owns a stream or can operate on itself.
	DEVICE = PermissionLevel(3)
	// FAMILY is for devices with the same owner, but no edit permissions
	FAMILY = PermissionLevel(2)
	// ENABLED is for any device that can do reading in the system
	ENABLED = PermissionLevel(1)
	// ANYBODY is for doing completely unpriviliged operations.
	ANYBODY = PermissionLevel(0)
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

// Gte checks that the given permission is at least what the desired one should be
func (actual PermissionLevel) Gte(desired PermissionLevel) bool {
	return uint(actual) >= uint(desired)
}

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

// Devices are general purposed external and internal data users,
//
type Device struct {
	DeviceId         int64  `modifiable:"nobody" json:"-"`                        // The primary key of this device
	Name             string `modifiable:"root" json:"name"`                       // The registered name of this device, should be universally unique like "Devicename_serialnum"
	Nickname         string `modifiable:"device" json:"nickname"`                 // The human readable name of this device
	UserId           int64  `modifiable:"root" json:"-"`                          // the user that owns this device
	ApiKey           string `modifiable:"device" json:"apikey,omitempty"`         // A uuid used as an api key to verify against
	Enabled          bool   `modifiable:"user" json:"enabled"`                    // Whether or not this device can do reading and writing
	IsAdmin          bool   `modifiable:"root" json:"admin,omitempty"`            // Whether or not this is a "superdevice" which has access to the whole API
	CanWrite         bool   `modifiable:"user" json:"canwrite,omitempty"`         // Can this device write to streams? (inactive right now)
	CanWriteAnywhere bool   `modifiable:"user" json:"canwriteanywhere,omitempty"` // Can this device write to others streams? (inactive right now)
	CanActAsUser     bool   `modifiable:"user" json:"user,omitempty"`             // Can this device operate as a user? (inactive right now)
	IsVisible        bool   `modifiable:"root" json:"visible"`
	UserEditable     bool   `modifiable:"root" json:"-"`
}

func (d *Device) ValidityCheck() error {
	if !IsValidName(d.Name) {
		return InvalidNameError
	}

	return nil
}

func (d *Device) GeneralPermissions() PermissionLevel {
	if !d.Enabled {
		return ANYBODY
	}

	if d.IsAdmin {
		return ROOT
	}

	return ENABLED
}

func (d *Device) RelationToUser(user *User) PermissionLevel {
	// guards
	if user == nil || !d.Enabled {
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

func (d *Device) RelationToDevice(device *Device) PermissionLevel {
	// guards
	if device == nil || !d.Enabled {
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

	return ENABLED
}

func (d *Device) RelationToStream(stream *Stream, streamParent *Device) PermissionLevel {
	// guards
	if stream == nil || streamParent == nil || !d.Enabled {
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

func (d *Device) RevertUneditableFields(originalValue Device, p PermissionLevel) int {
	return revertUneditableFields(reflect.ValueOf(d), reflect.ValueOf(originalValue), p)
}

func revertUneditableFields(toChange reflect.Value, originalValue reflect.Value, p PermissionLevel) int {

	//fmt.Printf("Getting original elem %v\n", originalValue.Kind())
	originalValueReflect := originalValue //.Elem()

	//fmt.Println("done getting elem")
	changeNumber := 0
	for i := 0; i < originalValueReflect.NumField(); i++ {
		// Grab the fields for reflection
		originalValueField := originalValueReflect.Field(i)
		typeField := originalValueReflect.Type().Field(i)

		// Check what kind of modifiable permission we need to edit
		modifiable := typeField.Tag.Get("modifiable")

		// By default, we don't allow modification
		if modifiable == "" {
			modifiable = "nobody"
		}

		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, originalValueField.Interface(), modifiable)

		//fmt.Printf("Field name: %v, modifiable %v\n",originalTypeField.Name, originalValueField.String(),  modifiable)

		// If we don't have enough permissions, reset the field from original
		requiredPermissionsForField, _ := strToPermissionLevel(modifiable)
		if !p.Gte(requiredPermissionsForField) {
			//fmt.Printf("Setting field\n")
			if !reflect.DeepEqual(toChange.Elem().Field(i).Interface(), originalValueField.Interface()) {
				toChange.Elem().Field(i).Set(originalValueField)
				changeNumber++
			}
		}
	}

	// and bob's your uncle!
	return changeNumber
}
