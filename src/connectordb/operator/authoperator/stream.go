package authoperator

import "connectordb/users"

//ReadAllStreamsByDeviceID reads all streams associated with the device
func (o *AuthOperator) ReadAllStreamsByDeviceID(deviceID int64) ([]users.Stream, error) {
	odev, err := o.Device()
	if err != nil {
		return nil, err
	}
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	if odev.RelationToDevice(dev).Gte(users.FAMILY) {
		return o.BaseOperator.ReadAllStreamsByDeviceID(deviceID)
	}
	return nil, err
}

//CreateStreamByDeviceID creates a new stream
func (o *AuthOperator) CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error {
	odev, err := o.Device()
	if err != nil {
		return err
	}
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return err
	}
	if odev.RelationToDevice(dev).Gte(users.DEVICE) {
		err = o.BaseOperator.CreateStreamByDeviceID(deviceID, streamname, jsonschema)
		if err == nil {
			devpath, err2 := o.getDevicePath(deviceID)
			if err2 == nil {
				o.MetaLog("CreateStream", devpath+"/"+streamname)
			}
		}
		return err
	}
	return ErrPermissions
}

/**
//ReadStream reads the given stream
func (o *AuthOperator) ReadStream(streampath string) (*users.Stream, error) {
	stream, err := o.Operator.ReadStream(streampath)
	if err != nil {
		return nil, err
	}

	if _, err := o.devPermissionsGteStream(stream, users.DEVICE); err != nil {
		return nil, err
	}

	return stream, nil
}
**/

//ReadStreamByID reads the given stream using its ID
func (o *AuthOperator) ReadStreamByID(streamID int64) (*users.Stream, error) {
	stream, err := o.BaseOperator.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	if _, err := o.devPermissionsGteStream(stream, users.DEVICE); err != nil {
		return nil, err
	}

	return stream, nil
}

//ReadStreamByDeviceID reads the stream given a device ID and the stream name
func (o *AuthOperator) ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error) {
	stream, err := o.BaseOperator.ReadStreamByDeviceID(deviceID, streamname)
	if err != nil {
		return nil, err
	}

	if _, err := o.devPermissionsGteStream(stream, users.DEVICE); err != nil {
		return nil, err
	}

	return stream, nil
}

//UpdateStream updates the stream
func (o *AuthOperator) UpdateStream(modifiedstream *users.Stream) error {
	originalStream, err := o.BaseOperator.ReadStreamByID(modifiedstream.StreamId)
	if err != nil {
		return err
	}

	streamsDevice, err := o.BaseOperator.ReadDeviceByID(originalStream.DeviceId)
	if err != nil {
		return err
	}

	myDevice, err := o.Device()
	if err != nil {
		return err
	}

	permission := myDevice.RelationToStream(originalStream, streamsDevice)

	if modifiedstream.RevertUneditableFields(*originalStream, permission) > 0 {
		return ErrPermissions
	}

	err = o.BaseOperator.UpdateStream(modifiedstream)
	if err == nil {
		o.MetaLogStreamID(originalStream.StreamId, "UpdateStream")
	}

	return err
}

/**
devPermissionsGteStream checks if this device's permissions are greater than or
equal to the level relative to the given stream.

Returns:

    PermissionLevel - the relation of the other user's device to this one,
					  nobody on error
	error - ErrPermissoins if the permission level is not set, or other errors
	        if a database issue occurred. nil if the relation permissionlevel
			is >= the requested one
**/
func (o *AuthOperator) devPermissionsGteStream(streamToCheck *users.Stream, permissionToCheck users.PermissionLevel) (users.PermissionLevel, error) {
	myDevice, err := o.Device()
	if err != nil {
		return users.NOBODY, err
	}

	streamsDevice, err := o.BaseOperator.ReadDeviceByID(streamToCheck.DeviceId)
	if err != nil {
		return users.NOBODY, err
	}

	permission := myDevice.RelationToStream(streamToCheck, streamsDevice)
	if permission.Gte(permissionToCheck) {
		return permission, nil
	}

	return users.NOBODY, ErrPermissions
}

//DeleteStreamByID Delete the stream using ID... This doesn't actually use the ID internally
func (o *AuthOperator) DeleteStreamByID(streamID int64, substream string) error {
	stream, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}

	if _, err := o.devPermissionsGteStream(stream, users.DEVICE); err != nil {
		return err
	}

	spath, err2 := o.getStreamPath(streamID)

	err = o.BaseOperator.DeleteStreamByID(streamID, substream)
	if err == nil && err2 == nil {
		o.MetaLog("DeleteStream", spath)
	}
	return err
}
