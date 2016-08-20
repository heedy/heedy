/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.

This file contains the functions for "devices". A device is maps to a real-world
device or service that can read a user's data and write to its streams.
**/
package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nu7hatch/gouuid"
)

// Devices are general purposed external and internal data users,
type Device struct {
	DeviceID    int64  `json:"-"`           // The primary key of this device
	Name        string `json:"name"`        // The registered name of this device, should be universally unique like "Devicename_serialnum"
	Nickname    string `json:"nickname"`    // The human readable name of this device
	Description string `json:"description"` // A public description
	Icon        string `json:"icon"`        // A public icon in a data URI format, should be smallish 100x100?
	UserID      int64  `json:"-"`           // the user that owns this device
	APIKey      string `json:"apikey"`      // A uuid used as an api key to verify against
	Enabled     bool   `json:"enabled"`     // Whether or not this device considers itself online (working/gathering)
	Public      bool   `json:"public"`      // Whether the device is accessible from public

	// The permissions allotted to this device
	Role string `json:"role"`

	IsVisible    bool `json:"visible"`
	UserEditable bool `json:"user_editable"`
}

// DeviceMaker is the structure used to create a device
type DeviceMaker struct {
	Device

	Streams map[string]*StreamMaker `json:"streams"`

	Devicelimit int64 `json:"-"`
}

// Validate ensures that the values are legal
func (dm *DeviceMaker) Validate(streamLimit int) error {
	if dm == nil {
		return errors.New("Null device creation struct")
	}

	err := dm.Device.ValidityCheck()
	if err != nil {
		return err
	}

	if streamLimit > 0 && len(dm.Streams) > streamLimit {
		return errors.New("Exceeded stream limit for device")
	}

	for s := range dm.Streams {
		dm.Streams[s].Name = s
		err = dm.Streams[s].Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) String() string {
	return fmt.Sprintf(`[users.Device |
	Id: %v,
	Name: %v,
	Nick: %v,
	Api: %v,
	Enabled: %v,
	Visible: %v,
	UserEdit: %v]`,
		d.DeviceID,
		d.Name,
		d.Nickname,
		d.APIKey,
		d.Enabled,
		d.IsVisible,
		d.UserEditable)
}

// ValidityCheck ensures valid name
func (d *Device) ValidityCheck() error {
	if !IsValidName(d.Name) {
		return InvalidNameError
	}

	return validateIcon(d.Icon)

}

// CreateDevice adds a device to the system given its owner and name.
// returns the last inserted id. It is assumed that the DeviceMaker was already validated.
// This means that DeviceMaker.Validate() had already been called, and returned nil
func (userdb *SqlUserDatabase) CreateDevice(d *DeviceMaker) error {
	APIKey, _ := uuid.NewV4()

	if d.Devicelimit > 0 {
		// TODO: This check should happen in a transaction, since the way it is done now enables timing attacks
		num, err := userdb.CountDevicesForUser(d.UserID)
		if err != nil {
			return err
		}
		if num >= d.Devicelimit {
			return errors.New("Can't create device: Device number limit exceeded.")
		}
	}

	_, err := userdb.Exec(`INSERT INTO devices
		(	name,
			apikey,
			userid,
			public,
			description,
			icon,
			nickname,
			enabled,
			role,
			isvisible,
			usereditable
		)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`, d.Name, APIKey.String(), d.UserID, d.Public,
		d.Description, d.Icon, d.Nickname, d.Enabled, d.Role, d.IsVisible, d.UserEditable)

	if err != nil && strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint ") {
		return errors.New("Device with this name already exists")
	}

	if len(d.Streams) > 0 {
		dev, err := userdb.ReadDeviceByAPIKey(APIKey.String())
		if err != nil {
			return err
		}
		for s := range d.Streams {
			d.Streams[s].Name = s
			d.Streams[s].DeviceID = dev.DeviceID
			if err = userdb.CreateStream(d.Streams[s]); err != nil {
				userdb.DeleteDevice(dev.DeviceID)
				return err
			}
		}
	}

	return err
}

// ReadDevicesForUserID gets all of a user's devices
func (userdb *SqlUserDatabase) ReadDevicesForUserID(UserID int64) ([]*Device, error) {
	var devices []*Device

	err := userdb.Select(&devices, "SELECT * FROM devices WHERE userid = ?;", UserID)

	if err == sql.ErrNoRows {
		return nil, ErrDeviceNotFound
	}

	return devices, err
}

// ReadDeviceForUserByName reads a device given a userID and device name
func (userdb *SqlUserDatabase) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	var dev Device

	err := userdb.Get(&dev, "SELECT * FROM devices WHERE userid = ? AND name = ? LIMIT 1;", userid, devicename)

	if err == sql.ErrNoRows {
		return nil, ErrDeviceNotFound
	}

	return &dev, err
}

// ReadDeviceByID selects the device with the given id from the database, returning nil if none can be found
func (userdb *SqlUserDatabase) ReadDeviceByID(DeviceID int64) (*Device, error) {
	var dev Device

	err := userdb.Get(&dev, "SELECT * FROM devices WHERE deviceid = ? LIMIT 1", DeviceID)

	if err == sql.ErrNoRows {
		return nil, ErrDeviceNotFound
	}

	return &dev, err

}

// ReadDeviceByAPIKey reads a device by an api key and returns it, it will be
// nil if an error was encountered and error will be set.
func (userdb *SqlUserDatabase) ReadDeviceByAPIKey(Key string) (*Device, error) {
	var dev Device

	if Key == "" {
		return nil, errors.New("Must have non-empty api key")
	}

	err := userdb.Get(&dev, "SELECT * FROM devices WHERE apikey = ? LIMIT 1;", Key)

	if err == sql.ErrNoRows {
		return nil, ErrDeviceNotFound
	}

	return &dev, err
}

// UpdateDevice updates the given device in the database with all fields in the
// struct.
func (userdb *SqlUserDatabase) UpdateDevice(device *Device) error {
	if device == nil {
		return InvalidPointerError
	}

	if err := device.ValidityCheck(); err != nil {
		return err
	}

	_, err := userdb.Exec(`UPDATE devices SET
		name = ?,
		nickname = ?,
		description = ?,
		icon = ?,
		userid = ?,
		apikey = ?,
		enabled = ?,
		role = ?,
		isvisible = ?,
		usereditable = ?,
		public = ? WHERE deviceid = ?;`,
		device.Name,
		device.Nickname,
		device.Description,
		device.Icon,
		device.UserID,
		device.APIKey,
		device.Enabled,
		device.Role,
		device.IsVisible,
		device.UserEditable,
		device.Public,
		device.DeviceID)

	return err
}

// DeleteDevice removes a device from the system.
func (userdb *SqlUserDatabase) DeleteDevice(ID int64) error {
	result, err := userdb.Exec(`DELETE FROM devices WHERE deviceid = ?;`, ID)
	return getDeleteError(result, err)
}
