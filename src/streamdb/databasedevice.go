package streamdb

import (
	"streamdb/users"
	"strings"

	"github.com/nu7hatch/gouuid"
)

//ReadAllDevices for the given user
func (o *Database) ReadAllDevices(username string) ([]users.Device, error) {
	u, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.Userdb.ReadDevicesForUserId(u.UserId)
}

//Technically, it is inefficient to pass in a path in a/b format, but our use case is
//so extremely dominated by database query/network, that it is essentially free to make stuff
//as pretty as possible.
func splitDevicePath(devicepath string) (usr string, dev string, err error) {
	splitted := strings.Split(devicepath, "/")
	if len(splitted) != 2 {
		return "", "", ErrBadPath
	}
	return splitted[0], splitted[1], nil
}

//ReadDeviceUser gets the user associated with the given device path
func (o *Database) ReadDeviceUser(devicepath string) (u *users.User, err error) {
	username, _, err := splitDevicePath(devicepath)
	if err != nil {
		return nil, err
	}
	return o.ReadUser(username)
}

//ReadUserAndDevice gets the user and device associated with the given path
func (o *Database) ReadUserAndDevice(devicepath string) (u *users.User, d *users.Device, err error) {
	username, _, err := splitDevicePath(devicepath)
	if err != nil {
		return nil, nil, err
	}
	usr, err := o.ReadUser(username)
	if err != nil {
		return nil, nil, err
	}
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, nil, err
	}
	if usr.UserId != dev.UserId {
		//We have an ID mismatch - the cache is outdated. Purge it, and try again
		o.userCache.Remove(username)
		o.deviceCache.Remove(devicepath)
		return o.ReadUserAndDevice(devicepath)
	}
	return usr, dev, nil
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

//ReadDeviceByID Note: Reading by ID cannot make use of the cache. It always touches the
//database. This makes sure that the ID is valid/fresh.
func (o *Database) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	dev, err := o.Userdb.ReadDeviceById(deviceID)

	//We can't save the device in cache, since we don't know the user name
	//and we don't want to waste another query to find it

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

//UpdateDevice updates the device at devicepath to the modifed device passed in
func (o *Database) UpdateDevice(devicepath string, modifieddevice *users.Device) error {
	username, devname, err := splitDevicePath(devicepath)
	if err != nil {
		return err
	}
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	if modifieddevice.RevertUneditableFields(*dev, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	err = o.Userdb.UpdateDevice(modifieddevice)
	if err == nil {
		//If the device name was changed, update device name in cache
		if devname != modifieddevice.Name {
			o.deviceCache.Remove(devicepath)
		}
		o.deviceCache.Add(username+"/"+modifieddevice.Name, *modifieddevice)
	}
	return err
}

//DeleteDevice deletes an existing device
func (o *Database) DeleteDevice(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	//Clean timebatchdb streams
	o.DeleteDeviceStreams(devicepath)

	err = o.Userdb.DeleteDevice(dev.DeviceId)
	o.deviceCache.Remove(devicepath)
	return err
}

//DeleteDeviceByID deletes the device using its deviceID
func (o *Database) DeleteDeviceByID(deviceID int64) error {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return err
	}
	usr, err := o.ReadUserByID(dev.UserId)
	if err != nil {
		return err
	}

	return o.DeleteDevice(usr.Name + "/" + dev.Name)
}

//ChangeDeviceAPIKey generates a new api key for the given device, and returns the key
func (o *Database) ChangeDeviceAPIKey(devicepath string) (apikey string, err error) {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return "", err
	}
	newkey, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	dev.ApiKey = newkey.String()
	return dev.ApiKey, o.UpdateDevice(devicepath, dev)
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
	//but only delete once from devices in postgres.

	//This function avoids the "user" device for good reason
	err = o.Userdb.DeleteAllDevicesForUser(usr.UserId)

	//Now loop through the devices, and delete them from cache if they exist
	//no need to worry about "user" here, since it can be reloaded
	for d := range devs {
		devpath := username + "/" + devs[d].Name
		o.DeleteDeviceStreams(devpath)
		o.deviceCache.Remove(devpath)
	}

	return err
}
