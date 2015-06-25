package users

import (
	"github.com/nu7hatch/gouuid"
)

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
