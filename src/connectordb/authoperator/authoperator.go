package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/operator"
	"connectordb/pathwrapper"
	"connectordb/users"

	pconfig "config/permissions"
)

// AuthOperator is the operator which represents actions as
// a particular logged in device
type AuthOperator struct {
	Operator operator.PathOperator
	pathwrapper.Wrapper

	devicePath string // The string name of this operator
	deviceID   int64  // The ID of this device
}

// NewAuthOperator creates a new authentication operator based upon the given DeviceID
func NewAuthOperator(op operator.PathOperator, deviceID int64) (*AuthOperator, error) {
	dev, err := op.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	usr, err := op.ReadUserByID(dev.UserID)
	if err != nil {
		return nil, err
	}

	ao := &AuthOperator{op, pathwrapper.Wrapper{}, usr.Name + "/" + dev.Name, deviceID}
	ao.Wrapper = pathwrapper.Wrap(ao)
	return ao, nil
}

// Name is the path to the device underlying the operator
func (a *AuthOperator) Name() string {
	return a.devicePath
}

// User returns the current user (ie, user that is logged in).
// No permissions checking is done
func (a *AuthOperator) User() (usr *users.User, err error) {
	dev, err := a.Operator.ReadDeviceByID(a.deviceID)
	if err != nil {
		return nil, err
	}
	return a.Operator.ReadUserByID(dev.UserID)
}

// Device returns the current device. No permissions checking
// is done on the device
func (a *AuthOperator) Device() (*users.Device, error) {
	return a.Operator.ReadDeviceByID(a.deviceID)
}

// AdminOperator returns the administrative operator
func (a *AuthOperator) AdminOperator() operator.PathOperator {
	return a.Operator.AdminOperator()
}

// getUserAndDevice returns both the current user AND the current device
// it is just there to simplify our work
func (a *AuthOperator) getUserAndDevice() (*users.User, *users.Device, error) {
	dev, err := a.Operator.ReadDeviceByID(a.deviceID)
	if err != nil {
		return nil, nil, err
	}
	u, err := a.Operator.ReadUserByID(dev.UserID)
	return u, dev, err
}

// getAccessLevels gets the access levels for the current user/device combo
func (a *AuthOperator) getAccessLevels(userID int64, ispublic, issself bool) (*pconfig.Permissions, *users.User, *users.Device, *pconfig.AccessLevel, *pconfig.AccessLevel, error) {
	u, d, err := a.getUserAndDevice()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	perm := pconfig.Get()

	up, dp := permissions.GetAccessLevels(perm, u, d, userID, ispublic, issself)
	return perm, u, d, up, dp, nil
}

// getDeviceAccessLevels is same as getAccessLevels, but it is given a deviceID
func (a *AuthOperator) getDeviceAccessLevels(deviceID int64) (*pconfig.Permissions, *users.Device, *users.User, *users.Device, *pconfig.AccessLevel, *pconfig.AccessLevel, error) {
	selfuser, selfdevice, err := a.getUserAndDevice()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	dev, err := a.Operator.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	perm := pconfig.Get()
	up, dp := permissions.GetAccessLevels(perm, selfuser, selfdevice, dev.UserID, dev.Public, selfdevice.DeviceID == dev.DeviceID)

	return perm, dev, selfuser, selfdevice, up, dp, nil
}
