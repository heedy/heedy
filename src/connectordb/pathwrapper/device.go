package pathwrapper

import (
	"connectordb/users"
	"util"
)

//ReadUserDevices returns all devices for the given user
func (w Wrapper) ReadUserDevices(username string) ([]*users.Device, error) {
	u, err := w.AdminOperator().ReadUser(username)
	if err != nil {
		return nil, err
	}
	return w.ReadAllDevicesByUserID(u.UserID)
}

//CreateDevice creates a new device at the given path
func (w Wrapper) CreateDevice(devicepath string) error {
	userName, deviceName, err := util.SplitDevicePath(devicepath)
	if err != nil {
		return err
	}
	u, err := w.AdminOperator().ReadUser(userName)
	if err != nil {
		return err
	}

	return w.CreateDeviceByUserID(u.UserID, deviceName)
}

// ReadDevice reads the given device
func (w Wrapper) ReadDevice(devicepath string) (*users.Device, error) {
	usrname, devname, err := util.SplitDevicePath(devicepath)
	if err != nil {
		return nil, err
	}
	u, err := w.AdminOperator().ReadUser(usrname)
	if err != nil {
		return nil, err
	}
	dev, err := w.ReadDeviceByUserID(u.UserID, devname)
	return dev, err
}

// UpdateDevice performs an update on the given device path
func (w Wrapper) UpdateDevice(devicepath string, updates map[string]interface{}) error {
	dev, err := w.AdminOperator().ReadDevice(devicepath)
	if err != nil {
		return err
	}
	return w.UpdateDeviceByID(dev.DeviceID, updates)
}

//DeleteDevice deletes an existing device
func (w Wrapper) DeleteDevice(devicepath string) error {
	dev, err := w.AdminOperator().ReadDevice(devicepath)
	if err != nil {
		return err
	}
	return w.DeleteDeviceByID(dev.DeviceID)
}
