package permissions

import (
	"config"
	"connectordb/users"
	"errors"

	log "github.com/Sirupsen/logrus"
)

// ErrNoAccess represents a total access error
var ErrNoAccess = errors.New("The device does not have access to this resource.")

// getPermissions gets the permissions level for the user
func getPermissions(cfg *config.Configuration, u *users.User) *config.Permissions {
	p, ok := cfg.Permissions[u.Permissions]
	if !ok {
		// The permissions level does not exist! Write an angry message to the console. This is a configuration error,
		// as such it should not be propagated to the user
		log.WithFields(log.Fields{"user": u.Name, "permissionslevel": u.Permissions}).Errorf("Could not find permissions level '%s'! Falling back to 'nobody'!", u.Permissions)
		return cfg.Permissions["nobody"]
	}
	return p
}

// ReadPublicAccessLevel gets the access level for a reading device for public data
func ReadPublicAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) (*config.AccessLevel, error) {
	if !d.CanReadExternal {
		return cfg.GetAccessLevel(cfg.Permissions["nobody"].PublicReadAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.GetAccessLevel(getPermissions(cfg, u).PublicReadAccessLevel)
}

// ReadPrivateAccessLevel gets the access level for a reading device for private data
func ReadPrivateAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) (*config.AccessLevel, error) {
	if !d.CanReadExternal {
		return cfg.GetAccessLevel(cfg.Permissions["nobody"].PrivateReadAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.GetAccessLevel(getPermissions(cfg, u).PrivateReadAccessLevel)
}

// ReadSelfAccessLevel gets the access level for a reading device for data about self
func ReadSelfAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) (*config.AccessLevel, error) {
	if !d.CanReadUser {
		return cfg.GetAccessLevel(cfg.Permissions["nobody"].SelfReadAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.GetAccessLevel(getPermissions(cfg, u).SelfReadAccessLevel)
}

// WritePublicAccessLevel gets the access level for a writing operation
func WritePublicAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) (*config.AccessLevel, error) {
	if !d.CanWriteExternal {
		return cfg.GetAccessLevel(cfg.Permissions["nobody"].PublicWriteAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.GetAccessLevel(getPermissions(cfg, u).PublicWriteAccessLevel)
}

// WritePrivateAccessLevel gets the access level for a writing operation
func WritePrivateAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) (*config.AccessLevel, error) {
	if !d.CanWriteExternal {
		return cfg.GetAccessLevel(cfg.Permissions["nobody"].PrivateWriteAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.GetAccessLevel(getPermissions(cfg, u).PrivateWriteAccessLevel)
}

// WriteSelfAccessLevel gets the access level for a writing operation
func WriteSelfAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) (*config.AccessLevel, error) {
	if !d.CanWriteUser {
		return cfg.GetAccessLevel(cfg.Permissions["nobody"].SelfWriteAccessLevel)
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.GetAccessLevel(getPermissions(cfg, u).SelfWriteAccessLevel)
}

// WriteOwnerAccessLevel gives the access level to the owning device
func WriteOwnerAccessLevel(cfg *config.Configuration, u *users.User) (*config.AccessLevel, error) {
	return cfg.GetAccessLevel(getPermissions(cfg, u).OwnerDeviceWriteAccessLevel)
}

// ReadOwnerAccessLevel gives the access level to the owning device
func ReadOwnerAccessLevel(cfg *config.Configuration, u *users.User) (*config.AccessLevel, error) {
	return cfg.GetAccessLevel(getPermissions(cfg, u).OwnerDeviceReadAccessLevel)
}
