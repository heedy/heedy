package connectordb

import (
	pconfig "config/permissions"
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"
)

// CountStreams returns the total nubmer of streams in the entire database
func (db *Database) CountStreams() (int64, error) {
	return db.Userdb.CountStreams()
}

// ReadAllStreamsByDeviceID reads all of a device's streams
func (db *Database) ReadAllStreamsByDeviceID(deviceID int64) ([]*users.Stream, error) {
	return db.Userdb.ReadStreamsByDevice(deviceID)
}

//CreateStreamByDeviceID creates the stream given a jsonschema as a string.
// It also enforces the max stream limit for the user
func (db *Database) CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error {
	perm := pconfig.Get()
	dev, err := db.ReadDeviceByID(deviceID)
	if err != nil {
		return err
	}
	u, err := db.ReadUserByID(dev.UserID)
	if err != nil {
		return err
	}

	r := permissions.GetUserRole(perm, u)

	return db.Userdb.CreateStream(streamname, jsonschema, deviceID, r.MaxStreams)
}

// ReadStreamByID reads the given stream
func (db *Database) ReadStreamByID(streamID int64) (*users.Stream, error) {
	return db.Userdb.ReadStreamByID(streamID)
}

// ReadStreamByDeviceID reads the given stream by its device ID and stream name
func (db *Database) ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error) {
	return db.Userdb.ReadStreamByDeviceIDAndName(deviceID, streamname)
}

// UpdateStreamByID updates the given stream
func (db *Database) UpdateStreamByID(streamID int64, updates map[string]interface{}) error {
	return errors.New("UNIMPLEMENTED")
}

// DeleteStreamByID removes the stream
func (db *Database) DeleteStreamByID(streamID int64, substream string) error {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return err //Workaround #81
	}

	if substream != "" {
		//We just delete the substream
		err = db.ds.DeleteSubstream(strm.DeviceID, strm.StreamID, substream)
	} else {
		err = db.Userdb.DeleteStream(streamID)
		if err == nil {
			err = db.ds.DeleteStream(strm.DeviceID, strm.StreamID)
		}
	}
	return err
}
