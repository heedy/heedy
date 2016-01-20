package permissions

import (
	"connectordb/users"
	"errors"
	"fmt"

	"config"
)

// ReadUser modifies the toread user taking away forbidden fields. It returns false if the user should not be visible
// to the reading device
func ReadUser(cfg *config.Configuration, readingUser *users.User, readingDevice *users.Device, toread *users.User) bool {
	var accessLevel *config.AccessLevel

	if toread.UserID == readingUser.UserID {
		accessLevel = ReadSelfAccessLevel(cfg, readingUser, readingDevice)
	} else if toread.Public {
		accessLevel = ReadPublicAccessLevel(cfg, readingUser, readingDevice)
	} else {
		accessLevel = ReadPrivateAccessLevel(cfg, readingUser, readingDevice)
	}

	if !accessLevel.CanAccessUser {
		return false
	}

	if !accessLevel.UserName {
		toread.Name = ""
	}
	if !accessLevel.UserNickname {
		toread.Nickname = ""
	}
	if !accessLevel.UserEmail {
		toread.Email = ""
	}
	if !accessLevel.UserDescription {
		toread.Description = ""
	}
	if !accessLevel.UserIcon {
		toread.Icon = ""
	}
	if !accessLevel.UserPermissions {
		toread.Permissions = ""
	}
	if !accessLevel.UserPublic {
		toread.Public = false
	}
	if !accessLevel.UserPassword {
		toread.Password = ""
	}

	return true
}

// WriteUser compares the original and modified user. It returns nil if the device can write the modifications, and an error if it cannot
// The modified user is to be a totally empty user (ie, not based on the actual user). This allows WriteUser to notice when fields were modified.
// The original user will have all of the changes integrated on success
func WriteUser(cfg *config.Configuration, writingUser *users.User, writingDevice *users.Device, original, modified *users.User) error {
	// We have to be careful while writing not to leak information about values
	var accessLevel *config.AccessLevel

	if original.UserID == writingUser.UserID {
		accessLevel = WriteSelfAccessLevel(cfg, writingUser, writingDevice)
	} else if original.Public {
		accessLevel = WritePublicAccessLevel(cfg, writingUser, writingDevice)
	} else {
		accessLevel = WritePrivateAccessLevel(cfg, writingUser, writingDevice)
	}

	if !accessLevel.CanAccessUser {
		return ErrNoAccess
	}

	if "" != modified.Name {
		if !accessLevel.UserName {
			return fmt.Errorf("This device does not have permissions necessary to write the name of user '%s'", original.Name)
		}
		return errors.New("ConnectorDB does not support modification of user names")
	}
	if "" != modified.Nickname {
		if !accessLevel.UserNickname {
			return fmt.Errorf("This device does not have permissions necessary to write the nickname of user '%s'", original.Name)
		}
		original.Nickname = modified.Nickname
	}
	if "" != modified.Email {
		if !accessLevel.UserEmail {
			return fmt.Errorf("This device does not have permissions necessary to write the email of user '%s'", original.Name)
		}
		original.Email = modified.Email
	}
	if "" != modified.Description {
		if !accessLevel.UserDescription {
			return fmt.Errorf("This device does not have permissions necessary to write the description of user '%s'", original.Name)
		}
		original.Description = modified.Description
	}
	if "" != modified.Icon {
		if !accessLevel.UserIcon {
			return fmt.Errorf("This device does not have permissions necessary to write the icon of user '%s'", original.Name)
		}
		original.Icon = modified.Icon
	}
	if "" != modified.Permissions {
		if !accessLevel.UserPermissions {
			return fmt.Errorf("This device does not have permissions necessary to write the permissions of user '%s'", original.Name)
		}
		// Check to make sure that the permissions exists
		_, ok := cfg.Permissions[modified.Permissions]
		if !ok {
			return fmt.Errorf("Permissions level '%s' does not exist.", modified.Permissions)
		}

		original.Permissions = modified.Permissions
	}

	if "" != modified.Password {
		if !accessLevel.UserPassword {
			return fmt.Errorf("This device does not have permissions necessary to write the password of user '%s'", original.Name)
		}
		// The password was modified - the password field must be hashed, so we do that here
		original.SetNewPassword(modified.Password)
	}

	// We have to be careful with booleans, since we can't tell if the user attempted modification and set to false, or didn't attempt
	// modification at all. We don't want to leak information.
	if accessLevel.UserPublic {
		original.Public = modified.Public
	} else if modified.Public {
		return fmt.Errorf("This device does not have permissions necessary to write the 'public' field of user '%s'", original.Name)
	}

	return nil
}
