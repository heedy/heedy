package connectordb

import (
	pconfig "config/permissions"
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"
	"fmt"

	"github.com/nu7hatch/gouuid"
)

func (db *Database) checkIfAddingDeviceWillExceedPrivateLimit(public bool, userID int64) (int64, int64, error) {
	var devs []*users.Device
	perm := pconfig.Get()
	u, err := db.ReadUserByID(userID)
	if err != nil {
		return 0, 0, err
	}
	if !u.Public && public {
		return 0, 0, errors.New("Can't make private user have public device")
	}

	r := permissions.GetUserRole(perm, u)

	// There are two devices by default: user and meta.
	// TODO: This should really be done all in a transaction on sql side, but we don't have time
	// to implement that now
	if r.MaxPrivateDevices > 0 && !public {
		devs, err = db.ReadAllDevicesByUserID(userID)
		if err != nil {
			return r.MaxDevices, r.MaxStreams, err
		}
		numprivate := int64(0)
		for i := range devs {
			if !devs[i].Public {
				numprivate++
			}
		}

		if numprivate >= r.MaxPrivateDevices {
			return r.MaxDevices, r.MaxStreams, errors.New("Exceeded maximum number of private devices for user")
		}

		// Just in case - Note that this is checked in userdb when creating a device!
		// we don't use >= because it is also used in updatedevice!
		if int64(len(devs)) > r.MaxDevices {
			return r.MaxDevices, r.MaxStreams, errors.New("Exceeded maximum number of devices for user")
		}
	}

	//
	return r.MaxDevices, r.MaxStreams, nil
}

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
func (db *Database) CreateDeviceByUserID(d *users.DeviceMaker) error {

	maxdev, maxstream, err := db.checkIfAddingDeviceWillExceedPrivateLimit(d.Public, d.UserID)
	if err != nil {
		return err
	}

	// Perform validation of the devicemaker query
	if err = d.Validate(int(maxstream)); err != nil {
		return err
	}
	d.Devicelimit = maxdev
	return db.Userdb.CreateDevice(d)
}

// ReadDeviceByID reads the given device
func (db *Database) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	return db.Userdb.ReadDeviceByID(deviceID)
}

// ReadDeviceByUserID reads a device given its user id and device name
func (db *Database) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	return db.Userdb.ReadDeviceForUserByName(userID, devicename)
}

// UpdateDeviceByID updates the device with the given map of update fields
func (db *Database) UpdateDeviceByID(deviceID int64, updates map[string]interface{}) error {

	d, err := db.ReadDeviceByID(deviceID)
	if err != nil {
		return err
	}

	oldname := d.Name
	waspublic := d.Public

	err = WriteObjectFromMap(d, updates)
	if err != nil {
		return err
	}

	if d.Name != oldname {
		return errors.New("ConnectorDB does not support modification of device names")
	}

	if !d.Public && waspublic {
		// Changing to private is same as adding a new private device
		_, _, err = db.checkIfAddingDeviceWillExceedPrivateLimit(false, d.UserID)
		if err != nil {
			return err
		}
	}
	if d.Role != "" {
		perm := pconfig.Get()
		_, ok := perm.DeviceRoles[d.Role]
		if !ok {
			return fmt.Errorf("Could not find device role '%s'", d.Role)
		}
	}

	if d.APIKey == "" {
		// Create a new API Key
		newkey, err := uuid.NewV4()
		if err != nil {
			// This should never happen...
			return fmt.Errorf("Failed to generate API Key: %s", err.Error())
		}
		d.APIKey = newkey.String()
	}

	return db.Userdb.UpdateDevice(d)
}

// DeleteDeviceByID deletes the given device
func (db *Database) DeleteDeviceByID(deviceID int64) error {
	err := db.Userdb.DeleteDevice(deviceID)
	if err == nil {
		err = db.DataStream.DeleteDevice(deviceID)
	}
	return err
}
