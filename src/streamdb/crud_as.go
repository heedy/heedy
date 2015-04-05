package streamdb

/**
This file provides the unified database public interface for the timebatchdb and
the users database.

If you want to connect to either of these, it is probably best to use this
package as it provides many conveniences.
**/

import (
	"errors"
	"streamdb/users"
	//"streamdb/timebatchdb"
	//"streamdb/dtypes"
)

type Permission int

const (
	USER           Permission = iota // the device is a user
	ACTIVE                           // The device is enabled
	ADMIN                            // The device is a superdevice (global superuser)
	WRITE                            // The device can write to user feeds
	WRITE_ANYWHERE                   // The device can write to any of a user's feeds
	MODIFY_USER                      // The device can modify it's owner
)

var (
	PrivilegeError            = errors.New("Insufficient privileges")
	InvalidParameterError     = errors.New("Invalid Parameter Recieved")
	super_privilege           = []Permission{ACTIVE, ADMIN}
	modify_user_privilege     = []Permission{ACTIVE, MODIFY_USER}
	active_privilege          = []Permission{ACTIVE}
	user_authorized_privilege = []Permission{ADMIN, USER, MODIFY_USER}
	write_privilege           = []Permission{WRITE, ACTIVE}
	write_anywhere_privilege  = []Permission{WRITE_ANYWHERE, ACTIVE}
	read_privilege            = []Permission{ACTIVE}
)

// Checks to see if the device has the listed permissions
func HasPermissions(d *users.Device, permissions []Permission) bool {
	for _, p := range permissions {
		switch p {
		case USER:
			if !d.CanActAsUser {
				return false
			}
		case ACTIVE:
			if !d.IsActive() {
				return false
			}
		case ADMIN:
			if !d.IsAdmin {
				return false
			}
		case WRITE:
			if !d.WriteAllowed() {
				return false
			}
		case WRITE_ANYWHERE:
			if !d.WriteAnywhereAllowed() {
				return false
			}
		case MODIFY_USER:
			if !d.CanActAsUser {
				return false
			}
		}
	}

	return true
}

func HasAnyPermission(d *users.Device, permissions []Permission) bool {
	for _, p := range permissions {
		switch p {
		case USER:
			if d.CanActAsUser {
				return true
			}
		case ACTIVE:
			if d.IsActive() {
				return true
			}
		case ADMIN:
			if d.IsAdmin {
				return true
			}
		case WRITE:
			if d.WriteAllowed() {
				return true
			}
		case WRITE_ANYWHERE:
			if d.WriteAnywhereAllowed() {
				return true
			}
		case MODIFY_USER:
			if d.CanActAsUser {
				return true
			}
		}
	}

	return false
}
