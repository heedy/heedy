package authoperator

import "connectordb/streamdb/users"

//ReadAllDevicesByUserID for the given user
func (o *AuthOperator) ReadAllDevicesByUserID(userID int64) ([]users.Device, error) {
	u, err := o.ReadUserByID(userID)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToUser(u).Gte(users.USER) {
		return o.Db.ReadAllDevicesByUserID(userID)
	}
	if usr, err := o.User(); err == nil && usr.UserId == u.UserId {
		return []users.Device{*dev}, err
	}
	return nil, ErrPermissions
}

//CreateDeviceByUserID creates a new device for the given user
func (o *AuthOperator) CreateDeviceByUserID(userID int64, devicename string) error {
	usr, err := o.Db.ReadUserByID(userID)
	if err != nil {
		return err
	}
	dev, err := o.Device()
	if err != nil {
		return err
	}
	if dev.RelationToUser(usr).Gte(users.USER) {
		err = o.Db.CreateDeviceByUserID(userID, devicename)
		if err == nil {
			o.UserLog("CreateDevice", usr.Name+"/"+devicename)
		}
	}
	return ErrPermissions
}

//ReadDevice reads the given device
func (o *AuthOperator) ReadDevice(devicepath string) (*users.Device, error) {
	readdev, err := o.Db.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToDevice(readdev).Gte(users.DEVICE) {
		return readdev, nil
	}
	return nil, ErrPermissions
}

//ReadDeviceByID reads the device using its ID
func (o *AuthOperator) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	readdev, err := o.Db.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToDevice(readdev).Gte(users.DEVICE) {
		return readdev, nil
	}
	return nil, ErrPermissions
}

//ReadDeviceByUserID reads the device using the user's ID and device name
func (o *AuthOperator) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	readdev, err := o.Db.ReadDeviceByUserID(userID, devicename)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToDevice(readdev).Gte(users.DEVICE) {
		return readdev, nil
	}
	return nil, ErrPermissions
}

//UpdateDevice updates the given device
func (o *AuthOperator) UpdateDevice(updateddevice *users.Device) error {
	dev, err := o.ReadDeviceByID(updateddevice.DeviceId)
	if err != nil {
		return err
	}
	operatordevice, err := o.Device()
	if err != nil {
		return err
	}
	permission := operatordevice.RelationToDevice(dev)
	if permission.Gte(users.DEVICE) && updateddevice.RevertUneditableFields(*dev, permission) == 0 {
		err = o.Db.UpdateDevice(updateddevice)
		if err == nil {
			o.UserLogDeviceID(dev.DeviceId, "UpdateDevice")
		}
	}
	return ErrPermissions
}

//DeleteDeviceByID deletes the device given its ID
func (o *AuthOperator) DeleteDeviceByID(deviceID int64) error {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return err
	}
	operatordevice, err := o.Device()
	if err != nil {
		return err
	}
	if operatordevice.RelationToDevice(dev).Gte(users.USER) {
		devpath, err2 := o.getDevicePath(deviceID)
		err = o.Db.DeleteDeviceByID(deviceID)
		if err == nil && err2 == nil {
			o.UserLog("DeleteDevice", devpath)
		}
	}
	return ErrPermissions
}
