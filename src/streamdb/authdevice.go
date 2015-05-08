package streamdb

import (
	"streamdb/users"

	"github.com/nu7hatch/gouuid"
)

//ReadDevice reads the given device
func (o *AuthOperator) ReadDevice(devicepath string) (*users.Device, error) {
	if o.Name() == devicepath {
		dev, err := o.Device()
		if err != nil {
			return nil, err
		}
		//getting device updates the name if it had changed
		if o.Name() == devicepath {
			return dev, nil
		}
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	newdevice, err := o.Db.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	if dev.RelationToDevice(newdevice).Gte(users.USER) {
		return newdevice, nil
	}
	return nil, ErrPermissions
}

//ReadDeviceByID reads the device using its ID
func (o *AuthOperator) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	newdevice, err := o.Db.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToDevice(newdevice).Gte(users.USER) {
		return newdevice, nil
	}
	return nil, ErrPermissions
}

//ReadAllDevices for the given user
func (o *AuthOperator) ReadAllDevices(username string) ([]users.Device, error) {
	u, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToUser(u).Gte(users.USER) {
		return o.Db.ReadAllDevices(username)
	}
	if o.usrName == username { //If this is the user, return this device
		usr, err := o.User()
		if err != nil {
			return nil, err
		}
		//Make sure that this user is valid with the reloaded name
		if usr.Name == username {
			return []users.Device{*dev}, err
		}
	}
	return nil, ErrPermissions
}

//CreateDevice creates a new device at the given path
func (o *AuthOperator) CreateDevice(devicepath string) error {
	usr, err := o.Db.ReadDeviceUser(devicepath)
	if err != nil {
		return err
	}
	dev, err := o.Device()
	if err != nil {
		return err
	}
	if dev.RelationToUser(usr).Gte(users.USER) {
		return o.Db.CreateDevice(devicepath)
	}
	return ErrPermissions
}

//UpdateDevice updates the given device
func (o *AuthOperator) UpdateDevice(devicepath string, updateddevice *users.Device) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	operatordevice, err := o.Device()
	if err != nil {
		return err
	}
	permission := operatordevice.RelationToDevice(dev)
	if permission.Gte(users.DEVICE) && updateddevice.RevertUneditableFields(*dev, permission) == 0 {
		return o.Db.UpdateDevice(devicepath, updateddevice)
	}
	return ErrPermissions
}

//DeleteDevice deletes an existing device
func (o *AuthOperator) DeleteDevice(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	operatordevice, err := o.Device()
	if err != nil {
		return err
	}
	if operatordevice.RelationToDevice(dev).Gte(users.USER) {
		return o.Db.DeleteDevice(devicepath)
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
		return o.Db.DeleteDeviceByID(deviceID)
	}
	return ErrPermissions
}

//ChangeDeviceAPIKey generates a new api key for the given device, and returns the key
func (o *AuthOperator) ChangeDeviceAPIKey(devicepath string) (apikey string, err error) {
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
