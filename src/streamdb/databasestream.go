package streamdb

import (
	"streamdb/schema"
	"streamdb/users"
	"strings"
)

//Technically, it is inefficient to pass in a path in a/b format, but our use case is
//so extremely dominated by database query/network, that it is essentially free to make stuff
//as pretty as possible.
func splitStreamPath(streampath string) (usr string, dev string, stream string, err error) {
	splitted := strings.Split(streampath, "/")
	if len(splitted) != 3 {
		return "", "", "", ErrBadPath
	}
	return splitted[0], splitted[1], splitted[2], nil
}

//ReadStreamDevice gets the device associated with the given stream path
func (o *Database) ReadStreamDevice(streampath string) (u *users.Device, err error) {
	username, devicename, _, err := splitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	return o.ReadDevice(username + "/" + devicename)
}

//CreateStream makes a new stream
func (o *Database) CreateStream(streampath, jsonschema string) error {

	//Validate that the schema is correct
	if _, err := schema.NewSchema(jsonschema); err != nil {
		return err
	}

	username, devicename, streamname, err := splitStreamPath(streampath)
	if err != nil {
		return err
	}
	dev, err := o.ReadDevice(username + "/" + devicename)
	if err != nil {
		return err
	}
	return o.Userdb.CreateStream(streamname, jsonschema, dev.DeviceId)
}

//ReadAllStreams reads all the streams for the given device
func (o *Database) ReadAllStreams(devicepath string) ([]Stream, error) {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	usrstrms, err := o.Userdb.ReadStreamsByDevice(dev.DeviceId)

	//Now convert the users.Stream to Stream objects
	strms := make([]Stream, len(usrstrms))
	for i := range usrstrms {
		strms[i], err = NewStream(&usrstrms[i], err)
	}
	return strms, err
}

//ReadStream reads the given stream
func (o *Database) ReadStream(streampath string) (*Stream, error) {
	//Check if the stream is in the cache
	if s, ok := o.streamCache.Get(streampath); ok {
		strm := s.(Stream)
		return &strm, nil
	}
	//Apparently not. Get the stream from userdb
	username, devicename, streamname, err := splitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	dev, err := o.ReadDevice(username + "/" + devicename)
	if err != nil {
		return nil, err
	}
	usrstrm, err := o.Userdb.ReadStreamByDeviceIdAndName(dev.DeviceId, streamname)
	strm, err := NewStream(usrstrm, err)
	if err != nil {
		return nil, err
	}

	o.streamCache.Add(streampath, strm) //This makes a copy in the cache
	return &strm, nil
}

//UpdateStream updates the stream. BUG(daniel) the function currently does not give an error
//if someone attempts to update the schema (which is an illegal operation anyways)
func (o *Database) UpdateStream(streampath string, modifiedstream *Stream) error {
	usrname, devname, streamname, err := splitStreamPath(streampath)
	if err != nil {
		return err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	if modifiedstream.RevertUneditableFields(strm.Stream, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	err = o.Userdb.UpdateStream(&modifiedstream.Stream)
	if err == nil {
		//If the stream name was changed, modify the stream name in cache
		if streamname != modifiedstream.Name {
			o.streamCache.Remove(streampath)
		}
		o.streamCache.Add(usrname+"/"+devname+"/"+modifiedstream.Name, *modifiedstream)
	}
	return err
}

//DeleteStream deletes the given stream
func (o *Database) DeleteStream(streampath string) error {
	s, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	err = o.Userdb.DeleteStream(s.StreamId)
	o.streamCache.Remove(streampath)
	return err
}

//DeleteDeviceStreams deletes all streams associated with the given device
func (o *Database) DeleteDeviceStreams(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	strms, err := o.ReadAllStreams(devicepath)
	if err != nil {
		return err
	}

	//Don't pound postgres
	err = o.Userdb.DeleteAllStreamsForDevice(dev.DeviceId)

	//Now loop through the streams, and delete them from cache if they exist
	for s := range strms {
		o.streamCache.Remove(devicepath + "/" + strms[s].Name)
		//TODO: Delete the data streams from timebatchdb
	}

	return err
}
