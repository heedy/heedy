package streamdb

import "streamdb/users"

//ReadStreamDevice gets the device associated with the given stream path
func (o *AuthOperator) ReadStreamDevice(streampath string) (d *users.Device, err error) {
	username, devicename, _, err := splitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	return o.ReadDevice(username + "/" + devicename)
}

//ReadStreamAndDevice reads both stream and device
func (o *AuthOperator) ReadStreamAndDevice(streampath string) (d *users.Device, s *Stream, err error) {
	username, devicename, _, err := splitStreamPath(streampath)
	if err != nil {
		return nil, nil, err
	}
	devicepath := username + "/" + devicename
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, nil, err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, nil, err
	}
	if strm.DeviceId != dev.DeviceId {
		o.Db.streamCache.Remove(streampath)
		o.Db.deviceCache.Remove(devicepath)
		return o.ReadStreamAndDevice(streampath)
	}
	return dev, strm, nil
}

//ReadStream reads the given stream
func (o *AuthOperator) ReadStream(streampath string) (*Stream, error) {
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	username, devicename, _, err := splitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	devicepath := username + "/" + devicename
	sdevice, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	strm, err := o.Db.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	if strm.DeviceId != dev.DeviceId {
		o.Db.streamCache.Remove(streampath)
		o.Db.deviceCache.Remove(devicepath)
		return o.ReadStream(streampath)
	}

	if dev.RelationToStream(&strm.Stream, sdevice).Gte(users.FAMILY) {
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
	sdevice, err := o.Db.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return nil, err
	}

	if dev.RelationToStream(&strm.Stream, sdevice).Gte(users.FAMILY) {
		return strm, nil
	}
	return nil, ErrPermissions
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

//UpdateStream updates the stream
func (o *AuthOperator) UpdateStream(streampath string, modifiedstream *Stream) error {
	odev, err := o.Device()
	if err != nil {
		return err
	}
	dev, strm, err := o.ReadStreamAndDevice(streampath)
	if err != nil {
		return err
	}
	permission := odev.RelationToStream(&strm.Stream, dev)
	if modifiedstream.RevertUneditableFields(strm.Stream, permission) > 0 {
		return ErrNotChangeable
	}
	return o.Db.UpdateStream(streampath, modifiedstream)
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
	stream, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}
	dev, err := o.ReadDeviceByID(stream.DeviceId)
	if err != nil {
		return err
	}
	usr, err := o.ReadUserByID(dev.UserId)
	if err != nil {
		return err
	}

	return o.DeleteStream(usr.Name + "/" + dev.Name + "/" + stream.Name)
}
