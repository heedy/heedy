package permissions

import (
	"connectordb/users"
	"errors"
	"fmt"

	pconfig "config/permissions"

	log "github.com/Sirupsen/logrus"
)

func copyMap(m map[string]bool) map[string]bool {
	cpy := make(map[string]bool)
	for k, v := range m {
		cpy[k] = v
	}
	return cpy
}

// GetUserWriteAccessLevel returns the access level necessary for writing. It requires the user and device that is doing the writing,
// and the UserID of the requested object (to check if users match)
func GetUserWriteAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device, userid int64, ispublic bool) map[string]bool {
	// We have to be careful while writing not to leak information about values
	var accessLevel *pconfig.AccessLevel
	var err error

	if u.UserID == userid {
		accessLevel, err = WriteSelfAccessLevel(cpm, u, d)
	} else if ispublic {
		accessLevel, err = WritePublicAccessLevel(cpm, u, d)
	} else {
		accessLevel, err = WritePrivateAccessLevel(cpm, u, d)
	}
	if err != nil {
		// The access level wasn't found: This is a configuration issue. This should never happen during runtime,
		// since configuration is validated before it is used. Nevertheless, to make it clear during testing/debugging we crash the program here.
		log.Fatal(err.Error())
	}

	amap := accessLevel.GetMap()

	return amap
}

func privilegeReadEscalator(self bool, amap map[string]bool) map[string]bool {
	amap = copyMap(amap) // Copy the map, since we don't want to change values in the underlying map
	if self {
		p := pconfig.Get().EscalatedReadSelfPrivileges
		for i := range p {
			amap[p[i]] = false
		}
	} else {
		p := pconfig.Get().EscalatedReadOtherPrivileges
		for i := range p {
			amap[p[i]] = false
		}
	}
}

func privilegeWriteEscalator(self bool, amap map[string]bool) map[string]bool {
	amap = copyMap(amap) // Copy the map, since we don't want to change values in the underlying map
	if self {
		p := pconfig.Get().EscalatedWriteSelfPrivileges
		for i := range p {
			amap[p[i]] = false
		}
	} else {
		p := pconfig.Get().EscalatedWriteOtherPrivileges
		for i := range p {
			amap[p[i]] = false
		}
	}
}

// GetUserReadAccessLevel returns the access level necessary for reading. It requires the user and device that is doing the reading,
// and the UserID of the requested object (to check if users match)
func GetUserReadAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device, userid int64, ispublic bool) map[string]bool {
	// We have to be careful while writing not to leak information about values
	var accessLevel *pconfig.AccessLevel
	var err error

	if u.UserID == userid {
		accessLevel, err = ReadSelfAccessLevel(cpm, u, d)
	} else if ispublic {
		accessLevel, err = ReadPublicAccessLevel(cpm, u, d)
	} else {
		accessLevel, err = ReadPrivateAccessLevel(cpm, u, d)
	}
	if err != nil {
		// The access level wasn't found: This is a configuration issue. This should never happen during runtime,
		// since configuration is validated before it is used. Nevertheless, to make it clear during testing/debugging we crash the program here.
		log.Fatal(err.Error())
	}
	return accessLevel.GetMap()
}

// ReadUserToMap returns a map with all readable fields. If it returns nil, it means that the user should not be accessible at
// all. The reason we use map[string]interface{} here as the output, is because we only want to include the readable fields.
// For example, if description: "" marshalled directly, then we don't know if we have permission to read it. To fix this,
// ReadUserToMap takes in a fill user, and returns a map with only the readable fields available, ready for json marshalling.
func ReadUserToMap(cpm *pconfig.Permissions, readingUser *users.User, readingDevice *users.Device, toread *users.User) map[string]interface{} {
	amap := GetUserReadAccessLevel(cpm, readingUser, readingDevice, toread.UserID, toread.Public)

	// To combat permissions escalation of devices we remove all privileges requiring escalation
	if !readingDevice.EscalatedPrivileges {
		amap = privilegeReadEscalator(readingUser.UserID == toread.UserID, amap)
	}

	if !amap["can_access_user"] {
		return ErrNoAccess
	}
	return ReadObjectToMap("user_", amap, toread)
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
func UpdateUserFromMap(cpm *pconfig.Permissions, writingUser *users.User, writingDevice *users.Device, original *users.User, modmap map[string]interface{}) error {
	amap := GetUserWriteAccessLevel(cpm, writingUser, writingDevice, original.UserID, original.Public)

	// To combat permissions escalation of devices we remove all privileges requiring escalation
	if !writingDevice.EscalatedPrivileges {
		amap = privilegeWriteEscalator(writingUser.UserID == original.UserID, amap)
	}

	if !amap["can_access_user"] {
		return ErrNoAccess
	}

	opassword := original.Password
	oname := original.Name
	operm := original.Role

	err := WriteObjectFromMap("user_", amap, original, modmap)
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
	_, ok := cpm.Role[original.Role]
	if operm != original.Role && !ok {
		return fmt.Errorf("Permissions level '%s' does not exist", original.Role)
	}

	return nil
}
