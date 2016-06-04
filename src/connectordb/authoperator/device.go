package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"
	"fmt"

	pconfig "config/permissions"
)

// CountDevices returns the total number of users of the entire database
func (a *AuthOperator) CountDevices() (int64, error) {
	perm := pconfig.Get()
	usr, dev, err := a.UserAndDevice()
	if err != nil {
		return 0, err
	}
	urole := permissions.GetUserRole(perm, usr)
	drole := permissions.GetDeviceRole(perm, dev)
	if !urole.CanCountDevices || !drole.CanCountDevices {
		return 0, errors.New("Don't have permissions necessary to count devices")
	}
	return a.Operator.CountDevices()
}

// ReadAllDevicesByUserID reads all of the devices belonging to this user which the authenticated device
// is allowed to read
func (a *AuthOperator) ReadAllDevicesByUserID(userID int64) ([]*users.Device, error) {
	u, err := a.Operator.ReadUserByID(userID)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	_, _, _, ua, da, err := a.getAccessLevels(userID, u.Public, false)
	if err != nil {
		return nil, err
	}
	if !ua.CanListDevices || !da.CanListDevices {
		return nil, permissions.ErrNoAccess
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

// ReadAllUsersToMap reads all of the users who this device has permissions to read to a map
func (a *AuthOperator) ReadUserDevicesToMap(uname string) ([]map[string]interface{}, error) {
	u, err := a.Operator.ReadUser(uname)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	_, _, _, ua, da, err := a.getAccessLevels(u.UserID, u.Public, false)
	if err != nil {
		return nil, err
	}
	if !ua.CanListDevices || !da.CanListDevices {
		return nil, permissions.ErrNoAccess
	}

	// See ReadAllUsers
	devs, err := a.Operator.ReadUserDevices(uname)
	result := make([]map[string]interface{}, 0, len(devs))
	for i := range devs {
		u, err := a.ReadDeviceToMap(uname + "/" + devs[i].Name)
		if err == nil {
			result = append(result, u)
		}
	}
	return result, nil
}

// DeviceMaker returns the DeviceMaker prepopulated with default values
// TODO: This is a hack - it does not set defaults for subdevices
// and substreams. Furthermore, create allows setting ALL properties,
// which is definitely not wanted
func (a *AuthOperator) DeviceMaker() (*users.DeviceMaker, error) {
	u, err := a.User()
	if err != nil {
		return nil, err
	}
	perm := pconfig.Get()
	// Make sure that the given role exists
	r, ok := perm.UserRoles[u.Role]
	if !ok {
		return nil, fmt.Errorf("The given role '%s' does not exist", u.Role)
	}

	d := r.CreateDeviceDefaults
	return &users.DeviceMaker{
		Device: users.Device{
			Nickname:     d.Nickname,
			Role:         d.Role,
			Description:  d.Description,
			Icon:         d.Icon,
			Public:       d.Public,
			IsVisible:    d.IsVisible,
			UserEditable: d.UserEditable,
			Enabled:      d.Enabled,
		},
	}, nil
}

// CreateDeviceByUserID attempts to create a device for the given user
func (a *AuthOperator) CreateDeviceByUserID(dm *users.DeviceMaker) error {
	u, err := a.Operator.ReadUserByID(dm.UserID)
	if err != nil {
		return permissions.ErrNoAccess
	}
	_, _, _, ua, da, err := a.getAccessLevels(dm.UserID, u.Public, false)
	if err != nil {
		return err
	}

	if !ua.CanCreateDevice || !da.CanCreateDevice {
		return permissions.ErrNoAccess
	}

	return a.Operator.CreateDeviceByUserID(dm)
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

// ReadDeviceToMap reads the given device into a map, where only the permitted fields are present in the map
func (a *AuthOperator) ReadDeviceToMap(devpath string) (map[string]interface{}, error) {
	dev, err := a.Operator.ReadDevice(devpath)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	perm, _, _, _, ua, da, err := a.getDeviceAccessLevels(dev.DeviceID)
	if err != nil {
		return nil, err
	}
	return permissions.ReadObjectToMap(perm, ua, da, "device", dev)
}

// ReadDeviceByUserID reads the given device by its name and user ID
func (a *AuthOperator) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	dev, err := a.Operator.ReadDeviceByUserID(userID, devicename)
	if err != nil {
		return nil, permissions.ErrNoAccess
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
		return permissions.ErrNoAccess
	}
	return a.Operator.DeleteDeviceByID(deviceID)
}
