package authoperator

import "connectordb/streamdb/users"

/**
devPermissionsGteDev checks if this device's permissions are greater than or
equal to the level relative to the given device.

Returns:

    PermissionLevel - the relation of the other user's device to this one.
	error - ErrPermissoins if the permission level is not set, or other errors
	        if a database issue occurred. nil if the relation permissionlevel
			is >= the requested one
**/
func (o *AuthOperator) permissionsGteDev(other *users.Device, level users.PermissionLevel) (users.PermissionLevel, error) {
	// Get AuthOperator's device
	dev, err := o.Device()
	if err != nil {
		return users.NOBODY, err
	}

	// Check if we have appropriate permissions
	permission := dev.RelationToDevice(other)
	if permission.Gte(level) {
		return permission, nil
	}

	return users.NOBODY, ErrPermissions
}

/**
permissionsGteUser checks if this device's permissions are greater than or
equal to the level relative to the given user.

Returns:

    PermissionLevel - the relation of the other user's device to this one.
	error - ErrPermissoins if the permission level is not set, or other errors
	        if a database issue occurred. nil if the relation permissionlevel
			is >= the requested one
**/
func (o *AuthOperator) permissionsGteUser(other *users.User, level users.PermissionLevel) (users.PermissionLevel, error) {
	// Get AuthOperator's device
	dev, err := o.Device()
	if err != nil {
		return users.NOBODY, err
	}

	// Check if we have appropriate permissions
	permission := dev.RelationToUser(other)
	if permission.Gte(level) {
		return permission, nil
	}

	return users.NOBODY, ErrPermissions
}

//ReadAllDevicesByUserID for the given user
func (o *AuthOperator) ReadAllDevicesByUserID(userID int64) ([]users.Device, error) {
	user, err := o.ReadUserByID(userID)
	if err != nil {
		return nil, err
	}

	permission, err := o.permissionsGteUser(user, users.FAMILY)
	if err != nil {
		return nil, err
	}

	if permission.Gte(users.USER) {
		return o.Operator.ReadAllDevicesByUserID(userID)
	}

	dev, err := o.Device()
	if err != nil {
		return nil, err
	}

	// We'll just return the current device.
	return []users.Device{*dev}, err
}

//CreateDeviceByUserID creates a new device for the given user
func (o *AuthOperator) CreateDeviceByUserID(userID int64, devicename string) error {
	user, err := o.Operator.ReadUserByID(userID)
	if err != nil {
		return err
	}

	if _, err := o.permissionsGteUser(user, users.USER); err != nil {
		return err
	}

	err = o.Operator.CreateDeviceByUserID(userID, devicename)
	if err == nil {
		o.UserLog("CreateDevice", user.Name+"/"+devicename)
	}

	return err
}

//ReadDevice reads the given device
func (o *AuthOperator) ReadDevice(devicepath string) (*users.Device, error) {
	dev, err := o.Operator.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}

	if _, err := o.permissionsGteDev(dev, users.DEVICE); err != nil {
		return nil, err
	}

	return dev, nil
}

// ReadDeviceByID reads the device using its ID
func (o *AuthOperator) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	dev, err := o.Operator.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}

	if _, err := o.permissionsGteDev(dev, users.DEVICE); err != nil {
		return nil, err
	}

	return dev, nil
}

// ReadDeviceByUserID reads the device using the user's ID and device name
func (o *AuthOperator) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	dev, err := o.Operator.ReadDeviceByUserID(userID, devicename)
	if err != nil {
		return nil, err
	}

	if _, err := o.permissionsGteDev(dev, users.DEVICE); err != nil {
		return nil, err
	}

	return dev, nil
}

// UpdateDevice updates the given device
func (o *AuthOperator) UpdateDevice(updateddevice *users.Device) error {
	dev, err := o.ReadDeviceByID(updateddevice.DeviceId)
	if err != nil {
		return err
	}

	permission, err := o.permissionsGteDev(dev, users.DEVICE)
	if err != nil {
		return err
	}

	if updateddevice.RevertUneditableFields(*dev, permission) != 0 {
		return ErrPermissions
	}

	err = o.Operator.UpdateDevice(updateddevice)
	if err == nil {
		o.UserLogDeviceID(dev.DeviceId, "UpdateDevice")
	}

	return err
}

//DeleteDeviceByID deletes the device given its ID
func (o *AuthOperator) DeleteDeviceByID(deviceID int64) error {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return err
	}

	if _, err := o.permissionsGteDev(dev, users.USER); err != nil {
		return err
	}

	devpath, err2 := o.getDevicePath(deviceID)
	err = o.Operator.DeleteDeviceByID(deviceID)
	if err == nil && err2 == nil {
		o.UserLog("DeleteDevice", devpath)
	}
	return err
}
