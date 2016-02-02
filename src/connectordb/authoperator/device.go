package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"

	pconfig "config/permissions"
)

// CountDevices returns the total number of users of the entire database
func (a *AuthOperator) CountDevices() (int64, error) {
	perm := pconfig.Get()
	usr, dev, err := a.getUserAndDevice()
	if err != nil {
		return 0, err
	}
	urole := permissions.GetUserRole(perm, usr)
	drole := permissions.GetDeviceRole(perm, dev)
	if !urole.CanCountDevices || !drole.CanCountDevices {
		return 0, errors.New("Don't have permissions necesaary to count devices")
	}
	return a.Operator.CountDevices()
}

// ReadAllDevicesByUserID reads all of the devices belonging to this user which the authenticated device
// is allowed to read
func (a *AuthOperator) ReadAllDevicesByUserID(userID int64) ([]*users.Device, error) {
	u, err := a.Operator.ReadUserByID(userID)
	if err != nil {
		return nil, err
	}
	_, _, _, ua, da, err := a.getAccessLevels(userID, u.Public, false)
	if err != nil {
		return nil, err
	}
	if !ua.CanListDevices || !da.CanListDevices {
		return nil, errors.New("You do not have permissions necessary to list this user's devices.")
	}

	// See ReadAllUsers
	devs, err := a.Operator.ReadAllDevicesByUserID(userID)
	if err != nil {
		return nil, err
	}
	result := make([]*users.Device, 0, len(devs))
	for i := range devs {
		d, err := a.ReadDeviceByID(devs[i].DeviceID)
		if err == nil {
			result = append(result, d)
		}
	}
	return result, nil
}

// CreateDeviceByUserID attempts to create a device for the given user
func (a *AuthOperator) CreateDeviceByUserID(userID int64, devicename string) error {
	u, err := a.Operator.ReadUserByID(userID)
	if err != nil {
		return err
	}
	_, _, _, ua, da, err := a.getAccessLevels(userID, u.Public, false)
	if err != nil {
		return err
	}

	if !ua.CanCreateDevice || !da.CanCreateDevice {
		return errors.New("You do not have permissions necessary to create this device.")
	}

	return a.Operator.CreateDeviceByUserID(userID, devicename)
}

// ReadDeviceByID reads the given device by ID
func (a *AuthOperator) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	perm, dev, _, _, ua, da, err := a.getDeviceAccessLevels(deviceID)
	if err != nil {
		return nil, err
	}

	err = permissions.DeleteDisallowedFields(perm, ua, da, "device", dev)
	if err != nil {
		return nil, err
	}

	return dev, nil
}

// ReadDeviceByUserID reads the given device by its name and user ID
func (a *AuthOperator) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	dev, err := a.Operator.ReadDeviceByUserID(userID, devicename)
	if err != nil {
		return nil, err
	}
	// Don't repeat code unnecessarily
	return a.ReadDeviceByID(dev.DeviceID)
}

// UpdateDeviceByID updates the device using its ID
func (a *AuthOperator) UpdateDeviceByID(deviceID int64, updates map[string]interface{}) error {
	perm, _, _, _, ua, da, err := a.getDeviceAccessLevels(deviceID)
	if err != nil {
		return err
	}

	err = permissions.CheckIfUpdateFieldsPermitted(perm, ua, da, "device", updates)
	if err != nil {
		return err
	}
	return a.Operator.UpdateDeviceByID(deviceID, updates)
}

// DeleteDeviceByID removes a device based upon its ID
func (a *AuthOperator) DeleteDeviceByID(deviceID int64) error {
	_, _, _, _, ua, da, err := a.getDeviceAccessLevels(deviceID)
	if err != nil {
		return err
	}
	if !ua.CanDeleteDevice || !da.CanDeleteDevice {
		return errors.New("You do not have permissions necessary to delete this device.")
	}
	return a.Operator.DeleteDeviceByID(deviceID)
}
