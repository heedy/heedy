package permissions

import (
	pconfig "config/permissions"
	"connectordb/users"

	log "github.com/Sirupsen/logrus"
)

// GetAccessLevels returns the two AccessLevel objects relevant to the query. The first is the user's access level, the second is the device's access level.
// This will never return an error, since ConectorDB has well-defined fallbacks (user-> nobody, and device -> none). If the roles of user/device
// are not found, it will return the fallbacks, and complain on the error log (error level)
func GetAccessLevels(perm *pconfig.Permissions, u *users.User, d *users.Device, queryUserID int64, ispublic, isself bool) (usr *pconfig.AccessLevel, dev *pconfig.AccessLevel) {
	userRole := GetUserRole(perm, u)
	deviceRole := GetDeviceRole(perm, d)

	// Now let's find out which access levels to return
	var userAccess string
	var deviceAccess string
	if isself {
		userAccess = userRole.SelfAccessLevel
		deviceAccess = deviceRole.SelfAccessLevel
	} else if queryUserID == u.UserID {
		userAccess = userRole.UserAccessLevel
		deviceAccess = deviceRole.UserAccessLevel
	} else if ispublic {
		userAccess = userRole.PublicAccessLevel
		deviceAccess = deviceRole.PublicAccessLevel
	} else {
		userAccess = userRole.PrivateAccessLevel
		deviceAccess = deviceRole.PrivateAccessLevel
	}

	// Now find these access levels. Note that we give a fatal error if they are not found, since
	// validation of configuration ensures that all levels that are mapped do indeed exist. If the
	// access levels are not found, it threfore means that there is some form of corruption, and we crash
	ua, err := perm.GetAccessLevel(userAccess)
	if err != nil {
		log.Fatalf("User Role '%s': %s", u.Role, err.Error())
	}
	da, err := perm.GetAccessLevel(deviceAccess)
	if err != nil {
		log.Fatalf("Device Role '%s': %s", d.Role, err.Error())
	}

	return ua, da
}

// GetReadAccess returns the RWAccess for read for the given access level.
func GetReadAccess(perm *pconfig.Permissions, accessLevel *pconfig.AccessLevel) *pconfig.RWAccess {
	rw, err := perm.GetRWAccess(accessLevel.ReadAccess)
	if err != nil {
		// The permissions configuration is corrupted
		log.Fatal(err.Error())
	}
	return rw
}

// GetWriteAccess returns the RWAccess for write for the given access level.
func GetWriteAccess(perm *pconfig.Permissions, accessLevel *pconfig.AccessLevel) *pconfig.RWAccess {
	rw, err := perm.GetRWAccess(accessLevel.WriteAccess)
	if err != nil {
		// The permissions configuration is corrupted
		log.Fatal(err.Error())
	}
	return rw
}
