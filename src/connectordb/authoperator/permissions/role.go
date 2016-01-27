package permissions

import (
	pconfig "config/permissions"
	"connectordb/users"

	log "github.com/Sirupsen/logrus"
)

// GetUserRole does not return an error, since we must ALWAYS have a valid role
// for any user, such that ConnectorDB knows what to do. Therefore, if the user's
// Role field is not found in the permissions configuration, an error is logged
// to console, but a "nobody" role is returned. This means that a user with wrong role
// will have same permissions as nobody (random site visitor).
// One thing to note: UserRole are not modifiable.
func GetUserRole(perm *pconfig.Permissions, u *users.User) *pconfig.UserRole {
	p, ok := perm.UserRoles[u.Role]
	if !ok {
		// The permissions level does not exist! Write an angry message to the console. This is a configuration error,
		// as such it should not be propagated to the user
		log.WithFields(log.Fields{"user": u.Name, "role": u.Role}).Errorf("Could not find user role '%s'! Falling back to 'nobody'!", u.Role)
		p, ok = perm.UserRoles["nobody"]
		if !ok {
			// OK, now THIS is a fatal error - the nobody role is REQUIRED - the configuration is corrupted
			log.Fatal("Could not find 'nobody' user role - configuration is corrupted!")
		}
	}
	return p
}

// GetDeviceRole behaves largely in the same way as GetUserRole.
func GetDeviceRole(perm *pconfig.Permissions, d *users.Device) *pconfig.DeviceRole {
	p, ok := perm.DeviceRoles[d.Role]
	if !ok {
		// The permissions level does not exist! Write an angry message to the console. This is a configuration error,
		// as such it should not be propagated to the user
		log.WithFields(log.Fields{"deviceid": d.DeviceID, "role": d.Role}).Errorf("Could not find device role '%s'! Falling back to 'none'!", d.Role)
		p, ok = perm.DeviceRoles["none"]
		if !ok {
			// OK, now THIS is a fatal error - the none role is REQUIRED - the configuration is corrupted
			log.Fatal("Could not find 'none' device role - configuration is corrupted!")
		}
	}
	return p
}
