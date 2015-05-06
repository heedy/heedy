package streamdb

import (
	"streamdb/users"
	"strings"
)

//ReadAllDevices for the given user
func (o *Database) ReadAllDevices(username string) ([]users.Device, error) {
	u, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.Userdb.ReadDevicesForUserId(u.UserId)
}

func splitDevicePath(devicepath string) (usr string, dev string, err error) {
	splitted := strings.Split(devicepath, "/")
	if len(splitted) != 2 {
		return "", "", ErrBadPath
	}
	return splitted[0], splitted[1], nil
}

//ReadDevice reads the given device
func (o *Database) ReadDevice(devicepath string) (*users.Device, error) {
	//Check if the device is in the cache
	if d, ok := o.deviceCache.Get(devicepath); ok {
		dev := d.(users.Device)
		return &dev, nil
	}
	//Apparently not. Get the device from userdb
	usrname, devname, err := splitDevicePath(devicepath)
	u, err := o.ReadUser(usrname)
	if err != nil {
		return nil, err
	}
	dev, err := o.Userdb.ReadDeviceForUserByName(u.UserId, devname)
	if err == nil {
		//Save the device in cache
		o.deviceCache.Add(devicepath, *dev)
	}

	return dev, err
}

//CreateDevice creates a new device at the given path
func (o *Database) CreateDevice(devicepath string) error {
	userName, deviceName, err := splitDevicePath(devicepath)
	if err != nil {
		return err
	}
	u, err := o.ReadUser(userName)
	if err != nil {
		return err
	}

	return o.Userdb.CreateDevice(deviceName, u.UserId)
}

//DeleteDevice deletes an existing device
func (o *Database) DeleteDevice(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	err = o.Userdb.DeleteDevice(dev.DeviceId)
	o.deviceCache.Remove(devicepath)
	return err
}

//DeleteUserDevices deletes all devices associated with the given user
func (o *Database) DeleteUserDevices(username string) error {
	usr, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	devs, err := o.ReadAllDevices(username)
	if err != nil {
		return err
	}

	//We will be more clever here than just running DeleteDevice in a loop
	//In particular, the whole goal here is to avoid pounding postgres, so
	//loop through deleting streams/devices from cache and timebatch,
	//but only delete once from devices.

	//This function avoids the "user" device for good reason
	err = o.Userdb.DeleteAllDevicesForUser(usr.UserId)

	//Now loop through the devices, and delete them from cache if they exist
	//no need to worry about "user" here, since it can be reloaded
	for d := range devs {
		o.deviceCache.Remove(username + "/" + devs[d].Name)
	}

	return err
}
