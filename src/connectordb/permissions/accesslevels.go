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
		// The permissions level does not exist! Write an angry message to the console
		log.WithFields(log.Fields{"user": u.Name, "permissions": u.Permissions}).Error("Could not find permissions level! Falling back to 'user'!")
		return cfg.Permissions["user"]
	}
	return p
}

// ReadPublicAccessLevel gets the access level for a reading device for public data
func ReadPublicAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) *config.AccessLevel {
	if !d.CanReadExternal {
		return cfg.AccessLevels[cfg.Permissions["nobody"].PublicReadAccessLevel]
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.AccessLevels[getPermissions(cfg, u).PublicReadAccessLevel]
}

// ReadPrivateAccessLevel gets the access level for a reading device for private data
func ReadPrivateAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) *config.AccessLevel {
	if !d.CanReadExternal {
		return cfg.AccessLevels[cfg.Permissions["nobody"].PrivateReadAccessLevel]
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.AccessLevels[getPermissions(cfg, u).PrivateReadAccessLevel]
}

// ReadSelfAccessLevel gets the access level for a reading device for data about self
func ReadSelfAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) *config.AccessLevel {
	if !d.CanReadUser {
		return cfg.AccessLevels[cfg.Permissions["nobody"].SelfReadAccessLevel]
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.AccessLevels[getPermissions(cfg, u).SelfReadAccessLevel]
}

// WritePublicAccessLevel gets the access level for a writing operation
func WritePublicAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) *config.AccessLevel {
	if !d.CanWriteExternal {
		return cfg.AccessLevels[cfg.Permissions["nobody"].PublicWriteAccessLevel]
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.AccessLevels[getPermissions(cfg, u).PublicWriteAccessLevel]
}

// WritePrivateAccessLevel gets the access level for a writing operation
func WritePrivateAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) *config.AccessLevel {
	if !d.CanWriteExternal {
		return cfg.AccessLevels[cfg.Permissions["nobody"].PrivateWriteAccessLevel]
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.AccessLevels[getPermissions(cfg, u).PrivateWriteAccessLevel]
}

// WriteSelfAccessLevel gets the access level for a writing operation
func WriteSelfAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device) *config.AccessLevel {
	if !d.CanWriteUser {
		return cfg.AccessLevels[cfg.Permissions["nobody"].SelfWriteAccessLevel]
	}
	// There can't be an error, since config is guaranteed to be validated
	return cfg.AccessLevels[getPermissions(cfg, u).SelfWriteAccessLevel]
}
