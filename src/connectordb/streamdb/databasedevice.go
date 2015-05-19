package streamdb

import (
	"connectordb/streamdb/users"
	"connectordb/streamdb/util"

	"github.com/nu7hatch/gouuid"
)

//ReadDeviceUser gets the user associated with the given device path
func (o *Database) ReadDeviceUser(devicepath string) (u *users.User, err error) {
	username, _, err := util.SplitDevicePath(devicepath, nil)
	if err != nil {
		return nil, err
	}
	return o.ReadUser(username)
}

//ReadAllDevices for the given user
func (o *Database) ReadAllDevices(username string) ([]users.Device, error) {
	u, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.ReadAllDevicesByUserID(u.UserId)
}

//ReadAllDevicesByUserID reads all devices for the given user's ID
func (o *Database) ReadAllDevicesByUserID(userID int64) ([]users.Device, error) {
	return o.Userdb.ReadDevicesForUserId(userID)
}

//CreateDevice creates a new device at the given path
func (o *Database) CreateDevice(devicepath string) error {
	userName, deviceName, err := util.SplitDevicePath(devicepath, nil)
	if err != nil {
		return err
	}
	u, err := o.ReadUser(userName)
	if err != nil {
		return err
	}

	return o.CreateDeviceByUserID(u.UserId, deviceName)
}

//CreateDeviceByUserID makes a new device using the UserID as source user
func (o *Database) CreateDeviceByUserID(userID int64, deviceName string) error {
	return o.Userdb.CreateDevice(deviceName, userID)
}

//ReadDevice reads the given device
func (o *Database) ReadDevice(devicepath string) (*users.Device, error) {
	//Check if the device is in the cache
	if d, ok := o.deviceCache.GetByName(devicepath); ok {
		dev := d.(users.Device)
		return &dev, nil
	}
	//Apparently not. Get the device from userdb
	usrname, devname, err := util.SplitDevicePath(devicepath, nil)
	if err != nil {
		return nil, err
	}
	u, err := o.ReadUser(usrname)
	if err != nil {
		return nil, err
	}
	dev, err := o.Userdb.ReadDeviceForUserByName(u.UserId, devname)
	if err == nil {
		//Save the device in cache
		o.deviceCache.Set(devicepath, dev.DeviceId, *dev)
	}

	return dev, err
}

//ReadDeviceByID Note: This version does not cache the path names, so querying by path to device
//will result in cache miss
func (o *Database) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	//Check if the device is in the cache
	if d, _, ok := o.deviceCache.GetByID(deviceID); ok {
		dev := d.(users.Device)
		return &dev, nil
	}

	dev, err := o.Userdb.ReadDeviceById(deviceID)

	if err == nil {
		//We add the device to the cache. But we don't know its full path, so see if the user is cached
		//to attempt recovery of username. If not, then just cache by ID alone
		if _, usrname, ok := o.userCache.GetByID(dev.UserId); ok {
			o.deviceCache.Set(usrname+"/"+dev.Name, dev.DeviceId, *dev)
		} else {
			o.deviceCache.SetID(dev.DeviceId, *dev)
		}
	}

	return dev, err
}

//ReadDeviceByUserID reads a device given a user's ID and the device name
func (o *Database) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	dev, err := o.Userdb.ReadDeviceForUserByName(userID, devicename)
	if err == nil {
		//TODO: Be more clever with finding the name here
		o.deviceCache.SetID(dev.DeviceId, *dev)
	}
	return dev, err
}

//UpdateDevice updates the device at devicepath to the modifed device passed in
func (o *Database) UpdateDevice(modifieddevice *users.Device) error {
	dev, err := o.ReadDeviceByID(modifieddevice.DeviceId)
	if err != nil {
		return err
	}
	if modifieddevice.RevertUneditableFields(*dev, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	err = o.Userdb.UpdateDevice(modifieddevice)

	//Now update the cache to the best of our ability
	if err == nil {
		if dev.Name == modifieddevice.Name {
			o.deviceCache.Update(dev.DeviceId, *modifieddevice) //Setting an ID
		} else {
			//Attempt to find the device path without database queries
			if _, usrname, _ := o.userCache.GetByID(dev.UserId); usrname != "" {
				o.deviceCache.Set(usrname+"/"+modifieddevice.Name, dev.DeviceId, *modifieddevice)
			} else {
				o.deviceCache.SetID(dev.DeviceId, *modifieddevice) //No luck with finding devpath
			}
		}
	}
	return err
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
	return dev.ApiKey, o.UpdateDevice(dev)
}

//DeleteDevice deletes an existing device
func (o *Database) DeleteDevice(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err //Workaround for #81
	}
	return o.DeleteDeviceByID(dev.DeviceId)
}

//DeleteDeviceByID deletes the device using its deviceID
func (o *Database) DeleteDeviceByID(deviceID int64) error {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return err //Workaround #81
	}

	//We read the user to clear the cache of the device's streams
	usr, err := o.ReadUserByID(dev.UserId)
	if err != nil {
		return err
	}

	err = o.Userdb.DeleteDevice(deviceID)
	o.deviceCache.RemoveID(deviceID)
	o.streamCache.UnlinkNamePrefix(usr.Name + "/" + dev.Name + "/")
	if err == nil {
		err = o.tdb.DeletePrefix(getTimebatchDeviceName(dev) + "/")
	}
	return err
}
