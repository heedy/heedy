package permissions

import (
	"connectordb/users"
	"errors"
	"fmt"

	"github.com/nu7hatch/gouuid"

	pconfig "config/permissions"

	log "github.com/Sirupsen/logrus"
)

// mergeMap merges two access level maps into one with greatest permissions
func mergeMap(a1 map[string]bool, a2 map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for key := range a1 {
		result[key] = a1[key] || a2[key]
	}
	return result
}

// GetDeviceReadAccessLevel gets the access level necessary to read the given device
func GetDeviceReadAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device, o *users.Device) (map[string]bool, error) {
	amap := GetUserReadAccessLevel(cpm, u, d, o.UserID, o.Public)
	// Now, we must merge this access level with the owner-specific one if this device is the accessing device
	if d.DeviceID == o.DeviceID {
		lvl, err := ReadOwnerAccessLevel(cpm, u)
		if err != nil {
			// See GetUserReadAccessLevel for explanation (this is not allowed to happen unless something is FUBAR)
			log.Fatal(err.Error())
		}
		amap = mergeMap(amap, lvl.GetMap())
	}

	// TODO: READ ESCALATOR. The issue here is how to encode the comparisons here

	if !d.EscalatedPrivileges {
		// There is a privilege escalation that we have to fix here: the device can't have API key access to any other devices
		// since it could just log in as those devices
		amap = copyMap(amap)
		amap["device_apikey"] = false
	}

	if !amap["can_access_device"] {
		return nil, ErrNoAccess
	}
	if !amap["can_access_non_user_editable_devices"] && !o.UserEditable {
		return nil, ErrNoAccess
	}
	return amap, nil
}

// GetDeviceWriteAccessLevel gets the access level necessary to write the given device
func GetDeviceWriteAccessLevel(cpm *pconfig.Permissions, u *users.User, d *users.Device, o *users.Device) (map[string]bool, error) {
	amap := GetUserWriteAccessLevel(cpm, u, d, o.UserID, o.Public)
	// Now, we must merge this access level with the owner-specific one if this device is the accessing device
	if d.DeviceID == o.DeviceID {
		lvl, err := WriteOwnerAccessLevel(cpm, u)
		if err != nil {
			// See GetUserReadAccessLevel for explanation (this is not allowed to happen unless something is FUBAR)
			log.Fatal(err.Error())
		}
		amap = mergeMap(amap, lvl.GetMap())

	}
	if !d.EscalatedPrivileges {
		amap = copyMap(amap)

		// We cannot allow the device to modify any device permissions to avoid privilege escalation
		amap["device_can_read_user"] = false
		amap["device_can_write_user"] = false
		amap["device_can_read_external"] = false
		amap["device_can_write_external"] = false
		amap["device_can_read_user_streams"] = false
		amap["device_can_write_user_streams"] = false
		amap["device_can_read_external_streams"] = false
		amap["device_can_write_external_streams"] = false
		amap["device_escalated_privileges"] = false

		// If the device is other, we cannot allow this device to change the API Key of the other
		// since it could then log in as that device
		if d.DeviceID != o.DeviceID {
			amap["device_apikey"] = false
		}
	}

	if !amap["can_access_device"] {
		return nil, ErrNoAccess
	}
	if !amap["can_access_non_user_editable_devices"] && !o.UserEditable {
		return nil, ErrNoAccess
	}
	return amap, nil
}

// ReadDeviceToMap : See ReadUserToMap
func ReadDeviceToMap(cpm *pconfig.Permissions, readingUser *users.User, readingDevice *users.Device, toread *users.Device) map[string]interface{} {
	amap, err := GetDeviceReadAccessLevel(cpm, readingUser, readingDevice, toread)
	if err != nil {
		return nil
	}

	return ReadObjectToMap("device_", amap, toread)
}

// UpdateDeviceFromMap : See UodateUserFromMap
func UpdateDeviceFromMap(cpm *pconfig.Permissions, writingUser *users.User, writingDevice *users.Device, original *users.Device, modmap map[string]interface{}) error {
	amap, err := GetDeviceWriteAccessLevel(cpm, writingUser, writingDevice, original)
	if err != nil {
		return err
	}

	oname := original.Name

	err = WriteObjectFromMap("device_", amap, original, modmap)
	if err != nil {
		return err
	}

	if oname != original.Name {
		return errors.New("ConnectorDB does not support modification of device names")
	}

	if original.Name == "user" {
		// The user device is special - it is REQUIRED that it have full user permissions
		if !original.CanReadUser || !original.CanReadExternal || !original.CanWriteUser || !original.CanWriteExternal ||
			!original.CanReadUserStreams || !original.CanReadExternalStreams || !original.CanWriteUserStreams || !original.CanWriteExternalStreams {
			return errors.New("The 'user' device must have full permissions")
		}
	}
	if original.APIKey == "" {
		// Generate a new api key if it is cleared
		newkey, err := uuid.NewV4()
		if err != nil {
			// This should never happen...
			return fmt.Errorf("Failed to generate API Key: %s", err.Error())
		}
		original.APIKey = newkey.String()

	}

	return nil
}
