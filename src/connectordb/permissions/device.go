package permissions

import (
	"connectordb/users"
	"errors"
	"fmt"

	"github.com/nu7hatch/gouuid"

	"config"

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
func GetDeviceReadAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device, o *users.Device) (map[string]bool, error) {
	amap := GetUserReadAccessLevel(cfg, u, d, o.UserID, o.Public).GetMap()
	// Now, we must merge this access level with the owner-specific one if this device is the accessing device
	if d.DeviceID == o.DeviceID {
		lvl, err := ReadOwnerAccessLevel(cfg, u)
		if err != nil {
			// See GetUserReadAccessLevel for explanation (this is not allowed to happen unless something is FUBAR)
			log.Fatal(err.Error())
		}
		amap = mergeMap(amap, lvl.GetMap())
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
func GetDeviceWriteAccessLevel(cfg *config.Configuration, u *users.User, d *users.Device, o *users.Device) (map[string]bool, error) {
	amap := GetUserWriteAccessLevel(cfg, u, d, o.UserID, o.Public).GetMap()
	// Now, we must merge this access level with the owner-specific one if this device is the accessing device
	if d.DeviceID == o.DeviceID {
		lvl, err := WriteOwnerAccessLevel(cfg, u)
		if err != nil {
			// See GetUserReadAccessLevel for explanation (this is not allowed to happen unless something is FUBAR)
			log.Fatal(err.Error())
		}
		amap = mergeMap(amap, lvl.GetMap())
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
func ReadDeviceToMap(cfg *config.Configuration, readingUser *users.User, readingDevice *users.Device, toread *users.Device) map[string]interface{} {
	amap, err := GetDeviceReadAccessLevel(cfg, readingUser, readingDevice, toread)
	if err != nil {
		return nil
	}

	return ReadObjectToMap("device_", amap, toread)
}

// UpdateDeviceFromMap : See UodateUserFromMap
func UpdateDeviceFromMap(cfg *config.Configuration, writingUser *users.User, writingDevice *users.Device, original *users.Device, modmap map[string]interface{}) error {
	amap, err := GetDeviceWriteAccessLevel(cfg, writingUser, writingDevice, original)
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
		if !original.CanReadUser || !original.CanReadExternal || !original.CanWriteUser || !original.CanWriteExternal {
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
