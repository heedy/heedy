package permissions

import (
	pconfig "config/permissions"
	"connectordb/users"
	"errors"

	log "github.com/Sirupsen/logrus"
)

// ErrNoAccess represents a total access error
var ErrNoAccess = errors.New("The device does not have access to this resource.")

// getRole gets the permissions level for the user
func getRole(cpm *pconfig.Permissions, u *users.User) *pconfig.Role {
	p, ok := cpm.Role[u.Role]
	if !ok {
		// The permissions level does not exist! Write an angry message to the console. This is a configuration error,
		// as such it should not be propagated to the user
		log.WithFields(log.Fields{"user": u.Name, "role": u.Role}).Errorf("Could not find role '%s'! Falling back to 'nobody'!", u.Role)
		return cpm.Role["nobody"]
	}
	return p
}

// ReadPublicAccessLevel gets the access level for a reading device for public data
func ReadPublicAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device) (*pconfig.AccessLevel, error) {
	if !d.CanReadExternal {
		return cpm.GetAccessLevel(cpm.Role["nobody"].PublicReadAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cpm.GetAccessLevel(getRole(cpm, u).PublicReadAccessLevel)
}

// ReadPrivateAccessLevel gets the access level for a reading device for private data
func ReadPrivateAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device) (*pconfig.AccessLevel, error) {
	if !d.CanReadExternal {
		return cpm.GetAccessLevel(cpm.Role["nobody"].PrivateReadAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cpm.GetAccessLevel(getRole(cpm, u).PrivateReadAccessLevel)
}

// ReadSelfAccessLevel gets the access level for a reading device for data about self
func ReadSelfAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device) (*pconfig.AccessLevel, error) {
	if !d.CanReadUser {
		return cpm.GetAccessLevel(cpm.Role["nobody"].SelfReadAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cpm.GetAccessLevel(getRole(cpm, u).SelfReadAccessLevel)
}

// WritePublicAccessLevel gets the access level for a writing operation
func WritePublicAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device) (*pconfig.AccessLevel, error) {
	if !d.CanWriteExternal {
		return cpm.GetAccessLevel(cpm.Role["nobody"].PublicWriteAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cpm.GetAccessLevel(getRole(cpm, u).PublicWriteAccessLevel)
}

// WritePrivateAccessLevel gets the access level for a writing operation
func WritePrivateAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device) (*pconfig.AccessLevel, error) {
	if !d.CanWriteExternal {
		return cpm.GetAccessLevel(cpm.Role["nobody"].PrivateWriteAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cpm.GetAccessLevel(getRole(cpm, u).PrivateWriteAccessLevel)
}

// WriteSelfAccessLevel gets the access level for a writing operation
func WriteSelfAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device) (*pconfig.AccessLevel, error) {
	if !d.CanWriteUser {
		return cpm.GetAccessLevel(cpm.Role["nobody"].SelfWriteAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cpm.GetAccessLevel(getRole(cpm, u).SelfWriteAccessLevel)
}

// WriteOwnerAccessLevel gives the access level to the owning device
func WriteOwnerAccessLevel(cpm *pconfig.Permissions, u *users.User) (*pconfig.AccessLevel, error) {
	return cpm.GetAccessLevel(getRole(cpm, u).OwnerDeviceWriteAccessLevel)
}

// ReadOwnerAccessLevel gives the access level to the owning device
func ReadOwnerAccessLevel(cpm *pconfig.Permissions, u *users.User) (*pconfig.AccessLevel, error) {
	return cpm.GetAccessLevel(getRole(cpm, u).OwnerDeviceReadAccessLevel)
}
