package permissions

import (
	"connectordb/users"
	"errors"

	"config"
)

// ReadStreamToMap : See ReadUserToMap
func ReadStreamToMap(cfg *config.Configuration, readingUser *users.User, readingDevice *users.Device, toreaddev *users.Device, toread *users.Stream) map[string]interface{} {
	amap, err := GetDeviceReadAccessLevel(cfg, readingUser, readingDevice, toreaddev)
	if err != nil {
		return nil
	}
	return ReadObjectToMap("stream_", amap, toread)
}

// UpdateDeviceFromMap : See UodateUserFromMap
func UpdateStreamFromMap(cfg *config.Configuration, writingUser *users.User, writingDevice *users.Device, originaldev *users.Device, original *users.Stream, modmap map[string]interface{}) error {
	amap, err := GetDeviceWriteAccessLevel(cfg, writingUser, writingDevice, originaldev)
	if err != nil {
		return err
	}
	oname := original.Name

	err = WriteObjectFromMap("stream_", amap, original, modmap)
	if err != nil {
		return err
	}

	if oname != original.Name {
		return errors.New("ConnectorDB does not support modification of stream names")
	}

	return nil
}
