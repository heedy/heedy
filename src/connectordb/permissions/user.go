package permissions

import (
	"connectordb/users"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"config"
)

// ReadUserToMap returns a map with all readable fields. If it returns nil, it means that the user should not be accessible at
// all. The reason we use map[string]interface{} here as the output, is because we only want to include the readable fields.
// For example, if description: "" marshalled directly, then we don't know if we have permission to read it. To fix this,
// ReadUserToMap takes in a fill user, and returns a map with only the readable fields available, ready for json marshalling.
func ReadUserToMap(cfg *config.Configuration, readingUser *users.User, readingDevice *users.Device, toread *users.User) map[string]interface{} {
	var accessLevel *config.AccessLevel
	var err error
	if toread.UserID == readingUser.UserID {
		accessLevel, err = ReadSelfAccessLevel(cfg, readingUser, readingDevice)
	} else if toread.Public {
		accessLevel, err = ReadPublicAccessLevel(cfg, readingUser, readingDevice)
	} else {
		accessLevel, err = ReadPrivateAccessLevel(cfg, readingUser, readingDevice)
	}
	if err != nil {
		// The access level wasn't found: This is a configuration issue. This should never happen during runtime,
		// since configuration is validated before it is used. Nevertheless, to make it clear we crash the program here.
		log.Fatal(err.Error())
	}
	if !accessLevel.CanAccessUser {
		return nil
	}
	return ReadObjectToMap("user_", accessLevel, toread)
}

// UpdateUserFromMap updates the "original" user object's fields to reflect the changes made in the modification map (modmap).
// The reason we can't really use a users.User object as our modification is because as the struct has values, we can't tell when
// the user attempted modification.
// This means that there are several "information leak" attacks if we use User objects:
//
// If we use an uninitialized user object, we cannot distinguish False booleans and empty strings from modification attempts, size
// those are the initial values.
//
// If we go by a modified user, then there is an information read leak - as an attacker, I can try reasonable values for a property,
// and keep retrying until I don't get an error. If I don't get an error, it means that the value I tried is the current value.
//
// Since I see no way of making modification work with the objects themselves, I chose to change to map[string]interface{} as the "output"
// type used in ConnectorDB
func UpdateUserFromMap(cfg *config.Configuration, writingUser *users.User, writingDevice *users.Device, original *users.User, modmap map[string]interface{}) error {
	// We have to be careful while writing not to leak information about values
	var accessLevel *config.AccessLevel
	var err error

	if original.UserID == writingUser.UserID {
		accessLevel, err = WriteSelfAccessLevel(cfg, writingUser, writingDevice)
	} else if original.Public {
		accessLevel, err = WritePublicAccessLevel(cfg, writingUser, writingDevice)
	} else {
		accessLevel, err = WritePrivateAccessLevel(cfg, writingUser, writingDevice)
	}
	if err != nil {
		// The access level wasn't found: This is a configuration issue. This should never happen during runtime,
		// since configuration is validated before it is used. Nevertheless, to make it clear we crash the program here.
		log.Fatal(err.Error())
	}

	if !accessLevel.CanAccessUser {
		return ErrNoAccess
	}

	opassword := original.Password
	oname := original.Name
	operm := original.Permissions

	err = WriteObjectFromMap("user_", accessLevel, original, modmap)
	if err != nil {
		return err
	}

	if opassword != original.Password {
		// The password needs to be set, since it involves multiple fields
		original.SetNewPassword(original.Password)
	}

	if oname != original.Name {
		return errors.New("ConnectorDB does not support modification of user names")
	}
	_, ok := cfg.Permissions[original.Permissions]
	if operm != original.Permissions && !ok {
		return fmt.Errorf("Permissions level '%s' does not exist", original.Permissions)
	}

	return nil
}
