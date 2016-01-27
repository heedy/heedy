package permissions

import (
	pconfig "config/permissions"
	"connectordb/users"
	"errors"
)

var (
	ErrNoPermissions = errors.New("This device does not have permissions necessary to perform the given action")
)

// CreateUser checks if the currently operating device has permissions necessary to create
// a user with the given data. Returns an error if permissions are not granted, and returns
// nil if everything is fine
func CreateUser(perm *pconfig.Permissions, u *users.User, d *users.Device, name, password, role string, public bool) error {
	// We get the permissions of device create/delete by looking at the corresponding public/private
	// AccessLevels. To get these, we set isself to false.
	ua, da := GetAccessLevels(perm, u, d, -1, public, false)

	if !ua.CanCreateUser || !da.CanCreateUser {
		return errors.New("You do not have permissions necessary to create a user.")
	}

	// Next, we check if the role is same as our own user's - if not, then we check if this device
	// has write permissions for roles.
	if u.Role != role {
		uw := GetWriteAccess(perm, ua)
		dw := GetWriteAccess(perm, da)
		if !uw.UserRole || !dw.UserRole {
			return errors.New("Don't have permission to create user with different role than creator")
		}
	}

	// Don't see a reason not to create the user
	return nil
}

// DeleteUser ensures that the currently operating device has the permissions necessary to delete the device

/*
func ReadUser(perm *pconfig.Permissions, u *users.User, d *users.Device, readme *users.User) (map[string]interface{}, error) {

}*/
