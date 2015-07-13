package streamdb

import (
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"
)

// ReadDeviceUser gets the user associated with the given device path
func (o *Database) ReadDeviceUser(devicepath string) (u *users.User, err error) {
	username, _, err := operator.SplitDevicePath(devicepath)
	if err != nil {
		return nil, err
	}
	return o.ReadUser(username)
}

// ReadAllDevicesByUserID reads all devices for the given user's ID
func (o *Database) ReadAllDevicesByUserID(userID int64) ([]users.Device, error) {
	return o.Userdb.ReadDevicesForUserId(userID)
}

// CreateDeviceByUserID makes a new device using the UserID as source user
func (o *Database) CreateDeviceByUserID(userID int64, deviceName string) error {
	return o.Userdb.CreateDevice(deviceName, userID)
}

// ReadDevice reads the given device
func (o *Database) ReadDevice(devicepath string) (*users.Device, error) {
	//Apparently not. Get the device from userdb
	usrname, devname, err := operator.SplitDevicePath(devicepath)
	if err != nil {
		return nil, err
	}
	u, err := o.ReadUser(usrname)
	if err != nil {
		return nil, err
	}
	dev, err := o.Userdb.ReadDeviceForUserByName(u.UserId, devname)
	return dev, err
}

// ReadDeviceByID Note: This version does not cache the path names, so querying by path to device
// will result in cache miss
func (o *Database) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	return o.Userdb.ReadDeviceById(deviceID)
}

// ReadDeviceByUserID reads a device given a user's ID and the device name
func (o *Database) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	return o.Userdb.ReadDeviceForUserByName(userID, devicename)
}

// UpdateDevice updates the device at devicepath to the modifed device passed in
func (o *Database) UpdateDevice(modifieddevice *users.Device) error {
	dev, err := o.ReadDeviceByID(modifieddevice.DeviceId)
	if err != nil {
		return err
	}
	if modifieddevice.RevertUneditableFields(*dev, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	return o.Userdb.UpdateDevice(modifieddevice)
}

// DeleteDeviceByID deletes the device using its deviceID
func (o *Database) DeleteDeviceByID(deviceID int64) error {

	err := o.Userdb.DeleteDevice(deviceID)
	if err == nil {
		err = o.ds.DeleteDevice(deviceID)
	}
	return err
}
