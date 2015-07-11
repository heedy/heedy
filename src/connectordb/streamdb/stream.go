package streamdb

import (
	"connectordb/streamdb/operator"
	"connectordb/streamdb/schema"
	"connectordb/streamdb/users"
)

//ReadAllStreamsByDeviceID reads all streams associated with the device with the given id
func (o *Database) ReadAllStreamsByDeviceID(deviceID int64) ([]operator.Stream, error) {
	usrstrms, err := o.Userdb.ReadStreamsByDevice(deviceID)

	//Now convert the users.Stream to Stream objects
	strms := make([]operator.Stream, len(usrstrms))
	for i := range usrstrms {
		strms[i], err = operator.NewStream(&usrstrms[i], err)
	}
	return strms, err
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
func (o *Database) ReadStream(streampath string) (*operator.Stream, error) {
	//Make sure that substreams are stripped from read
	_, devicepath, streampath, streamname, _, err := operator.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}

	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	usrstrm, err := o.Userdb.ReadStreamByDeviceIdAndName(dev.DeviceId, streamname)
	strm, err := operator.NewStream(usrstrm, err)
	if err != nil {
		return nil, err
	}

	//Now we add the stream to cache
	return &strm, nil
}

//ReadStreamByID reads a stream using a stream's ID
func (o *Database) ReadStreamByID(streamID int64) (*operator.Stream, error) {
	usrstrm, err := o.Userdb.ReadStreamById(streamID)
	strm, err := operator.NewStream(usrstrm, err)
	if err != nil {
		return nil, err
	}

	return &strm, err
}

//ReadStreamByDeviceID reads a stream given its name and the ID of its parent device
func (o *Database) ReadStreamByDeviceID(deviceID int64, streamname string) (*operator.Stream, error) {
	usrstrm, err := o.Userdb.ReadStreamByDeviceIdAndName(deviceID, streamname)
	strm, err := operator.NewStream(usrstrm, err)
	if err != nil {
		return nil, err
	}
	return &strm, nil
}

//UpdateStream updates the stream. BUG(daniel) the function currently does not give an error
//if someone attempts to update the schema (which is an illegal operation anyways)
func (o *Database) UpdateStream(modifiedstream *operator.Stream) error {
	strm, err := o.ReadStreamByID(modifiedstream.StreamId)
	if err != nil {
		return err
	}
	if modifiedstream.RevertUneditableFields(strm.Stream, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	err = o.Userdb.UpdateStream(&modifiedstream.Stream)
	if err == nil {
		if strm.Downlink == true && modifiedstream.Downlink == false {
			//There was a downlink here. Since the downlink was removed, we delete the associated
			//downlink substream
			o.DeleteStreamByID(strm.StreamId, "downlink")
		}
	}
	return err
}

//DeleteStreamByID deletes the stream using ID
func (o *Database) DeleteStreamByID(streamID int64, substream string) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err //Workaround #81
	}

	if substream != "" {
		//We just delete the substream
		err = o.ds.DeleteSubstream(strm.DeviceId, strm.StreamId, substream)
	} else {
		err = o.Userdb.DeleteStream(streamID)
		if err == nil {
			err = o.ds.DeleteStream(strm.DeviceId, strm.StreamId)
		}
	}
	return err

}
