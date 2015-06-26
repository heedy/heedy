package users

import (
	"reflect"

	"github.com/nu7hatch/gouuid"
)

// Devices are general purposed external and internal data users,
//
type Device struct {
	DeviceId         int64  `modifiable:"nobody" json:"-"`                        // The primary key of this device
	Name             string `modifiable:"user" json:"name"`                       // The registered name of this device, should be universally unique like "Devicename_serialnum"
	Nickname         string `modifiable:"device" json:"nickname"`                 // The human readable name of this device
	UserId           int64  `modifiable:"root" json:"-"`                          // the user that owns this device
	ApiKey           string `modifiable:"device" json:"apikey,omitempty"`         // A uuid used as an api key to verify against
	Enabled          bool   `modifiable:"user" json:"enabled"`                    // Whether or not this device can do reading and writing
	IsAdmin          bool   `modifiable:"root" json:"admin,omitempty"`            // Whether or not this is a "superdevice" which has access to the whole API
	CanWrite         bool   `modifiable:"user" json:"canwrite,omitempty"`         // Can this device write to streams? (inactive right now)
	CanWriteAnywhere bool   `modifiable:"user" json:"canwriteanywhere,omitempty"` // Can this device write to others streams? (inactive right now)
	CanActAsUser     bool   `modifiable:"user" json:"user,omitempty"`             // Can this device operate as a user? (inactive right now)
	IsVisible        bool   `modifiable:"root" json:"visible"`
	UserEditable     bool   `modifiable:"root" json:"-"`
}

func (d *Device) ValidityCheck() error {
	if !IsValidName(d.Name) {
		return InvalidNameError
	}

	return nil
}

func (d *Device) GeneralPermissions() PermissionLevel {
	if !d.Enabled {
		return ANYBODY
	}

	if d.IsAdmin {
		return ROOT
	}

	return ENABLED
}

func (d *Device) RelationToUser(user *User) PermissionLevel {
	// guards
	if user == nil || !d.Enabled {
		return ANYBODY
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT
	}

	if d.UserId == user.UserId {
		if d.CanActAsUser {
			return USER
		}

		return DEVICE
	}

	return ANYBODY
}

func (d *Device) RelationToDevice(device *Device) PermissionLevel {
	// guards
	if device == nil || !d.Enabled {
		return ANYBODY
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT
	}

	if d.UserId == device.UserId {
		if d.CanActAsUser {
			return USER
		}

		if d.DeviceId == device.DeviceId {
			return DEVICE
		}

		return FAMILY
	}

	return ENABLED
}

func (d *Device) RelationToStream(stream *Stream, streamParent *Device) PermissionLevel {
	// guards
	if stream == nil || streamParent == nil || !d.Enabled {
		return ANYBODY
	}

	// Permision Levels
	if d.IsAdmin {
		return ROOT
	}

	if d.CanActAsUser && d.UserId == streamParent.UserId {
		return USER
	}

	if d.DeviceId == stream.DeviceId {
		return DEVICE
	}

	if d.UserId == streamParent.UserId {
		return FAMILY
	}

	return ENABLED
}

func (d *Device) RevertUneditableFields(originalValue Device, p PermissionLevel) int {
	return revertUneditableFields(reflect.ValueOf(d), reflect.ValueOf(originalValue), p)
}

// CreateDevice adds a device to the system given its owner and name.
// returns the last inserted id
func (userdb *SqlUserDatabase) CreateDevice(Name string, UserId int64) error {
	ApiKey, _ := uuid.NewV4()

	if !IsValidName(Name) {
		return InvalidNameError
	}

	_, err := userdb.Exec(`INSERT INTO Devices
	    (	Name,
	        ApiKey,
	        UserId)
	        VALUES (?,?,?)`, Name, ApiKey.String(), UserId)

	return err
}

func (userdb *SqlUserDatabase) ReadDevicesForUserId(UserId int64) ([]Device, error) {
	var devices []Device

	err := userdb.Select(&devices, "SELECT * FROM Devices WHERE UserId = ?;", UserId)

	return devices, err
}

func (userdb *SqlUserDatabase) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	var dev Device

	err := userdb.Get(&dev, "SELECT * FROM Devices WHERE UserId = ? AND Name = ? LIMIT 1;", userid, devicename)

	return &dev, err
}

// ReadDeviceById selects the device with the given id from the database, returning nil if none can be found
func (userdb *SqlUserDatabase) ReadDeviceById(DeviceId int64) (*Device, error) {
	var dev Device

	err := userdb.Get(&dev, "SELECT * FROM Devices WHERE DeviceId = ? LIMIT 1", DeviceId)

	return &dev, err

}

// ReadDeviceByApiKey reads a device by an api key and returns it, it will be
// nil if an error was encountered and error will be set.
func (userdb *SqlUserDatabase) ReadDeviceByApiKey(Key string) (*Device, error) {
	var dev Device

	err := userdb.Get(&dev, "SELECT * FROM Devices WHERE ApiKey = ? LIMIT 1;", Key)

	return &dev, err
}

// UpdateDevice updates the given device in the database with all fields in the
// struct.
func (userdb *SqlUserDatabase) UpdateDevice(device *Device) error {
	if device == nil {
		return ERR_INVALID_PTR
	}

	if err := device.ValidityCheck(); err != nil {
		return err
	}

	_, err := userdb.Exec(`UPDATE Devices SET
	    Name = ?,
		Nickname = ?,
		UserId = ?,
		ApiKey = ?,
		Enabled = ?,
		IsAdmin = ?,
		CanWrite = ?,
		CanWriteAnywhere = ?,
		CanActAsUser = ?,
		IsVisible = ?,
		UserEditable = ? WHERE DeviceId = ?;`,
		device.Name,
		device.Nickname,
		device.UserId,
		device.ApiKey,
		device.Enabled,
		device.IsAdmin,
		device.CanWrite,
		device.CanWriteAnywhere,
		device.CanActAsUser,
		device.IsVisible,
		device.UserEditable,
		device.DeviceId)

	return err
}

// DeleteDevice removes a device from the system.
func (userdb *SqlUserDatabase) DeleteDevice(Id int64) error {
	_, err := userdb.Exec(`DELETE FROM Devices WHERE DeviceId = ?;`, Id)
	return err
}
