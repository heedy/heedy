package streamdb

import (
	"streamdb/users"
	"streamdb/util"
)

//ReadStreamDevice gets the device associated with the given stream path
func (o *AuthOperator) ReadStreamDevice(streampath string) (d *users.Device, err error) {
	_, devicepath, _, _, _, err := util.SplitStreamPath(streampath, nil)
	if err != nil {
		return nil, err
	}
	return o.ReadDevice(devicepath)
}

//ReadStreamAndDevice reads both stream and device
func (o *AuthOperator) ReadStreamAndDevice(streampath string) (d *users.Device, s *Stream, err error) {
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, nil, err
	}
	dev, err := o.ReadDeviceByID(strm.DeviceId)
	return dev, strm, err
}

//ReadAllStreams reads all streams associated with the device
func (o *AuthOperator) ReadAllStreams(devicepath string) ([]Stream, error) {
	odev, err := o.Device()
	if err != nil {
		return nil, err
	}
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	if odev.RelationToDevice(dev).Gte(users.FAMILY) {
		return o.Db.ReadAllStreams(devicepath)
	}
	return nil, err
}

//ReadAllStreamsByDeviceID reads all streams associated with the device
func (o *AuthOperator) ReadAllStreamsByDeviceID(deviceID int64) ([]Stream, error) {
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

//CreateStream creates a new stream
func (o *AuthOperator) CreateStream(streampath, jsonschema string) error {
	odev, err := o.Device()
	if err != nil {
		return err
	}
	dev, err := o.ReadStreamDevice(streampath)
	if err != nil {
		return err
	}
	if odev.RelationToDevice(dev).Gte(users.DEVICE) {
		return o.Db.CreateStream(streampath, jsonschema)
	}
	return ErrPermissions
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
		return o.Db.CreateStreamByDeviceID(deviceID, streamname, jsonschema)
	}
	return ErrPermissions
}

//ReadStream reads the given stream
func (o *AuthOperator) ReadStream(streampath string) (*Stream, error) {
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
func (o *AuthOperator) ReadStreamByID(streamID int64) (*Stream, error) {
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
func (o *AuthOperator) ReadStreamByDeviceID(deviceID int64, streamname string) (*Stream, error) {
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
func (o *AuthOperator) UpdateStream(modifiedstream *Stream) error {
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
		return ErrNotChangeable
	}
	return o.Db.UpdateStream(modifiedstream)
}

//DeleteStream deletes the stream at the given path
func (o *AuthOperator) DeleteStream(streampath string) error {
	odev, err := o.Device()
	if err != nil {
		return err
	}
	dev, strm, err := o.ReadStreamAndDevice(streampath)
	if err != nil {
		return err
	}
	if odev.RelationToStream(&strm.Stream, dev).Gte(users.DEVICE) {
		return o.Db.DeleteStream(streampath)
	}
	return ErrPermissions
}

//DeleteStreamByID Delete the stream using ID... This doesn't actually use the ID internally
func (o *AuthOperator) DeleteStreamByID(streamID int64) error {
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
		return o.Db.DeleteStreamByID(streamID)
	}
	return ErrPermissions
}
