package authoperator

import (
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"
)

//ReadStreamDevice gets the device associated with the given stream path
func (o *AuthOperator) ReadStreamDevice(streampath string) (d *users.Device, err error) {
	_, devicepath, _, _, _, err := operator.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	return o.ReadDevice(devicepath)
}

//ReadStreamAndDevice reads both stream and device
func (o *AuthOperator) ReadStreamAndDevice(streampath string) (d *users.Device, s *operator.Stream, err error) {
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, nil, err
	}
	dev, err := o.ReadDeviceByID(strm.DeviceId)
	return dev, strm, err
}

//ReadAllStreamsByDeviceID reads all streams associated with the device
func (o *AuthOperator) ReadAllStreamsByDeviceID(deviceID int64) ([]operator.Stream, error) {
	odev, err := o.Device()
	if err != nil {
		return nil, err
	}
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	if odev.RelationToDevice(dev).Gte(users.FAMILY) {
		return o.Db.ReadAllStreamsByDeviceID(deviceID)
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
		err = o.Db.CreateStreamByDeviceID(deviceID, streamname, jsonschema)
		if err == nil {
			devpath, err2 := o.getDevicePath(deviceID)
			if err2 == nil {
				o.UserLog("CreateStream", devpath+"/"+streamname)
			}
		}
		return err
	}
	return ErrPermissions
}

//ReadStream reads the given stream
func (o *AuthOperator) ReadStream(streampath string) (*operator.Stream, error) {
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	strm, err := o.Db.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	sdevice, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return nil, err
	}

	if dev.RelationToStream(&strm.Stream, sdevice).Gte(users.DEVICE) {
		return strm, nil
	}
	return nil, ErrPermissions
}

//ReadStreamByID reads the given stream using its ID
func (o *AuthOperator) ReadStreamByID(streamID int64) (*operator.Stream, error) {
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	strm, err := o.Db.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	sdevice, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return nil, err
	}

	if dev.RelationToStream(&strm.Stream, sdevice).Gte(users.DEVICE) {
		return strm, nil
	}
	return nil, ErrPermissions
}

//ReadStreamByDeviceID reads the stream given a device ID and the stream name
func (o *AuthOperator) ReadStreamByDeviceID(deviceID int64, streamname string) (*operator.Stream, error) {
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	strm, err := o.Db.ReadStreamByDeviceID(deviceID, streamname)
	if err != nil {
		return nil, err
	}
	sdevice, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return nil, err
	}

	if dev.RelationToStream(&strm.Stream, sdevice).Gte(users.DEVICE) {
		return strm, nil
	}
	return nil, ErrPermissions
}

//UpdateStream updates the stream
func (o *AuthOperator) UpdateStream(modifiedstream *operator.Stream) error {
	odev, err := o.Device()
	if err != nil {
		return err
	}
	strm, err := o.ReadStreamByID(modifiedstream.StreamId)
	if err != nil {
		return err
	}
	dev, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return err
	}
	permission := odev.RelationToStream(&strm.Stream, dev)
	if modifiedstream.RevertUneditableFields(strm.Stream, permission) > 0 {
		return ErrPermissions
	}
	err = o.Db.UpdateStream(modifiedstream)
	if err == nil {
		o.UserLogStreamID(strm.StreamId, "UpdateStream")
	}
	return err
}

//DeleteStreamByID Delete the stream using ID... This doesn't actually use the ID internally
func (o *AuthOperator) DeleteStreamByID(streamID int64, substream string) error {
	odev, err := o.Device()
	if err != nil {
		return err
	}
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}
	dev, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return err
	}
	if odev.RelationToStream(&strm.Stream, dev).Gte(users.DEVICE) {

		spath, err2 := o.getStreamPath(streamID)

		err = o.Db.DeleteStreamByID(streamID, substream)
		if err == nil && err2 == nil {
			o.UserLog("DeleteStream", spath)
		}
		return err
	}
	return ErrPermissions
}
