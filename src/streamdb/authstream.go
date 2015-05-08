package streamdb

import "streamdb/users"

//ReadStreamDevice gets the device associated with the given stream path
func (o *AuthOperator) ReadStreamDevice(streampath string) (u *users.Device, err error) {
	username, devicename, _, err := splitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	return o.ReadDevice(username + "/" + devicename)
}

//ReadStream reads the given stream
func (o *AuthOperator) ReadStream(streampath string) (*Stream, error) {
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	sdevice, err := o.ReadStreamDevice(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := o.Db.ReadStream(streampath)
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
	dev, err := o.ReadStreamDevice(streampath)
	if err != nil {
		return err
	}
	strm, err := o.ReadStream(streampath)
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
	dev, err := o.ReadStreamDevice(streampath)
	if err != nil {
		return err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	if odev.RelationToStream(&strm.Stream, dev).Gte(users.DEVICE) {
		return o.Db.DeleteStream(streampath)
	}
	return ErrPermissions
}
