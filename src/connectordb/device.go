package connectordb

import (
	pconfig "config/permissions"
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"
)

// CountDevices returns the total nubmer of devices in the entire database
func (db *Database) CountDevices() (int64, error) {
	return db.Userdb.CountDevices()
}

// ReadAllDevicesByUserID returns all devices that belong to the given user
func (db *Database) ReadAllDevicesByUserID(userID int64) ([]*users.Device, error) {
	return db.Userdb.ReadDevicesForUserID(userID)
}

// CreateDeviceByUserID creates a new device for the given user. It ensures that the permitted number
// of devices is not exceeded
func (db *Database) CreateDeviceByUserID(userID int64, devicename string) error {
	perm := pconfig.Get()
	u, err := db.ReadUserByID(userID)
	if err != nil {
		return err
	}

	r := permissions.GetUserRole(perm, u)

	// There are two devices by default: user and meta.
	// TODO: This should really be done all in a transaction on sql side, but we don't have time
	// to implement that now
	if r.MaxPrivateDevices > 2 {
		devs, err := db.ReadAllDevicesByUserID(userID)
		if err != nil {
			return err
		}
		numprivate := int64(0)
		for i := range devs {
			if !devs[i].Public {
				numprivate++
			}
		}

		if numprivate >= r.MaxPrivateDevices {
			return errors.New("Exceeded maximum number of private devices")
		}

		if int64(len(devs)) > r.MaxDevices {
			return errors.New("Exceeded maximum number of devices")
		}
	}

	return db.Userdb.CreateDevice(devicename, userID, r.MaxDevices)
}

// ReadDeviceByID reads the given device
func (db *Database) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	return db.Userdb.ReadDeviceByID(deviceID)
}

// ReadDeviceByUserID reads a device given its user id and device name
func (db *Database) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	return db.Userdb.ReadDeviceForUserByName(userID, devicename)
}

// ReadDeviceByAPIKey reads a device using only its api key
func (db *Database) ReadDeviceByAPIKey(apikey string) (*users.Device, error) {
	return db.Userdb.ReadDeviceByAPIKey(apikey)
}

// UpdateDeviceByID updates the device with the given map of update fields
func (db *Database) UpdateDeviceByID(deviceID int64, updates map[string]interface{}) error {
	return errors.New("UNIMPLEMENTED")
}

// DeleteDeviceByID deletes the given device
func (db *Database) DeleteDeviceByID(deviceID int64) error {
	err := db.Userdb.DeleteDevice(deviceID)
	if err == nil {
		err = db.ds.DeleteDevice(deviceID)
	}
	return err
}
