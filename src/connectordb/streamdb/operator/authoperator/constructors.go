package authoperator

import (
	"connectordb/streamdb/operator/interfaces"
	"connectordb/streamdb/util"
)

func NewUserAuthOperator(baseOperator interfaces.Operator, username string) (interfaces.BaseOperator, error) {
	devicePath := username + "/user"
	return NewDeviceAuthOperator(baseOperator, devicePath)
}

func NewDeviceAuthOperator(baseOperator interfaces.Operator, devicepath string) (interfaces.BaseOperator, error) {
	path, err := util.CreatePath(devicepath)
	if !path.IsDevice() || err != nil {
		return interfaces.ErrOperator{}, ErrBadPath
	}

	device, err := baseOperator.ReadDevice(devicepath)
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	userLogStream, err := baseOperator.ReadStream(path.User + "/user/log")
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	return &AuthOperator{baseOperator, devicepath, device.DeviceId, userLogStream.StreamId}, nil
}

func NewAPILoginOperator(baseOperator interfaces.Operator, apikey string) (interfaces.BaseOperator, error) {
	device, err := baseOperator.ReadDeviceByAPIKey(apikey)
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	return NewDeviceIdOperator(baseOperator, device.DeviceId)
}

func NewDeviceLoginOperator(baseOperator interfaces.Operator, devicepath, apikey string) (interfaces.BaseOperator, error) {
	operator, err := NewDeviceAuthOperator(baseOperator, devicepath)

	device, err := baseOperator.ReadDevice(devicepath)
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	if device.ApiKey != apikey {
		return interfaces.ErrOperator{}, ErrPermissions
	}

	return operator, nil
}

func NewDeviceIdOperator(baseOperator interfaces.Operator, deviceID int64) (interfaces.BaseOperator, error) {
	device, err := baseOperator.ReadDeviceByID(deviceID)
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	user, err := baseOperator.ReadUserByID(device.UserId)
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	userLogStream, err := baseOperator.ReadStream(user.Name + "/user/log")
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	devicepath := user.Name + "/" + device.Name

	return &AuthOperator{baseOperator, devicepath, deviceID, userLogStream.StreamId}, nil
}

func NewUserLoginOperator(baseOperator interfaces.Operator, username, password string) (interfaces.BaseOperator, error) {

	user, device, err := baseOperator.Login(username, password)
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	userLogStream, err := baseOperator.ReadStream(user.Name + "/user/log")
	if err != nil {
		return interfaces.ErrOperator{}, err
	}

	devicepath := user.Name + "/" + device.Name

	return &AuthOperator{baseOperator, devicepath, device.DeviceId, userLogStream.StreamId}, nil
}
