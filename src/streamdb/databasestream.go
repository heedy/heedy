package streamdb

import (
	"errors"
	"streamdb/schema"
	"streamdb/users"
	"streamdb/util"
)

var (
	ErrUnimplemented = errors.New("This operation is not yet implemented")
)

//ReadAllStreams reads all the streams for the given device
func (o *Database) ReadAllStreams(devicepath string) ([]Stream, error) {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	return o.ReadAllStreamsByDeviceID(dev.DeviceId)
}

//ReadAllStreamsByDeviceID reads all streams associated with the device with the given id
func (o *Database) ReadAllStreamsByDeviceID(deviceID int64) ([]Stream, error) {
	usrstrms, err := o.Userdb.ReadStreamsByDevice(deviceID)

	//Now convert the users.Stream to Stream objects
	strms := make([]Stream, len(usrstrms))
	for i := range usrstrms {
		strms[i], err = NewStream(&usrstrms[i], err)
	}
	return strms, err
}

//CreateStream makes a new stream
func (o *Database) CreateStream(streampath, jsonschema string) error {
	_, devicepath, _, streamname, substreams, err := util.SplitStreamPath(streampath, nil)
	if err != nil {
		return err
	}
	if substreams != "" {
		return ErrUnimplemented
	}

	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	return o.CreateStreamByDeviceID(dev.DeviceId, streamname, jsonschema)
}

//CreateStreamByDeviceID creates a stream using a device ID instead of path
func (o *Database) CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error {
	//Validate that the schema is correct
	if _, err := schema.NewSchema(jsonschema); err != nil {
		return err
	}
	return o.Userdb.CreateStream(streamname, jsonschema, deviceID)
}

//ReadStream reads the given stream
func (o *Database) ReadStream(streampath string) (*Stream, error) {
	//Check if the stream is in the cache
	if s, ok := o.streamCache.GetByName(streampath); ok {
		strm := s.(Stream)
		return &strm, nil
	}
	//Apparently not. Get the stream from userdb
	_, devicepath, _, streamname, substreams, err := util.SplitStreamPath(streampath, nil)
	if err != nil {
		return nil, err
	}
	if substreams != "" {
		return nil, ErrUnimplemented
	}
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	usrstrm, err := o.Userdb.ReadStreamByDeviceIdAndName(dev.DeviceId, streamname)
	strm, err := NewStream(usrstrm, err)
	if err != nil {
		return nil, err
	}

	//Now we add the stream to cache
	o.streamCache.Set(streampath, strm.StreamId, strm) //This makes a copy in the cache
	return &strm, nil
}

//ReadStreamByID reads a stream using a stream's ID
func (o *Database) ReadStreamByID(streamID int64) (*Stream, error) {
	if s, _, ok := o.streamCache.GetByID(streamID); ok {
		strm := s.(Stream)
		return &strm, nil
	}

	usrstrm, err := o.Userdb.ReadStreamById(streamID)
	strm, err := NewStream(usrstrm, err)
	if err != nil {
		return nil, err
	}

	//Add the stream to the cache. Since we don't know its full path, see if its device is cached,
	//and attempt to take the path from there
	if _, devpath, ok := o.deviceCache.GetByID(strm.DeviceId); ok && devpath != "" {
		o.streamCache.Set(devpath+"/"+strm.Name, strm.StreamId, strm)
	} else {
		o.streamCache.SetID(strm.StreamId, strm)
	}

	return &strm, err
}

//ReadStreamByDeviceID reads a stream given its name and the ID of its parent device
func (o *Database) ReadStreamByDeviceID(deviceID int64, streamname string) (*Stream, error) {
	usrstrm, err := o.Userdb.ReadStreamByDeviceIdAndName(deviceID, streamname)
	strm, err := NewStream(usrstrm, err)
	if err != nil {
		return nil, err
	}
	o.streamCache.SetID(strm.StreamId, strm)
	return &strm, nil
}

//UpdateStream updates the stream. BUG(daniel) the function currently does not give an error
//if someone attempts to update the schema (which is an illegal operation anyways)
func (o *Database) UpdateStream(modifiedstream *Stream) error {
	strm, err := o.ReadStreamByID(modifiedstream.StreamId)
	if err != nil {
		return err
	}
	if modifiedstream.RevertUneditableFields(strm.Stream, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	err = o.Userdb.UpdateStream(&modifiedstream.Stream)
	if err == nil {
		//If the stream name was changed, modify the stream name in cache
		if strm.Name == modifiedstream.Name {
			o.streamCache.Update(strm.StreamId, *modifiedstream)
		} else {
			//Attempt to recover the path by name using only cache
			if _, devpath, _ := o.deviceCache.GetByID(strm.DeviceId); devpath != "" {
				o.streamCache.Set(devpath+"/"+modifiedstream.Name, strm.StreamId, *modifiedstream)
			} else {
				o.streamCache.SetID(strm.StreamId, *modifiedstream)
			}
		}
	}
	return err
}

//DeleteStream deletes the given stream
func (o *Database) DeleteStream(streampath string) error {
	s, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	return o.DeleteStreamByID(s.StreamId)
}

//DeleteStreamByID deletes the stream using ID
func (o *Database) DeleteStreamByID(streamID int64) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err //Workaround #81
	}

	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return err
	}
	err = o.Userdb.DeleteStream(streamID)
	if err == nil {
		err = o.tdb.Delete(sname)
	}
	o.streamCache.RemoveID(streamID)
	return err

}
