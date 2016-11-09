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
func (db *Database) CreateStreamByDeviceID(s *users.StreamMaker) error {
	perm := pconfig.Get()
	dev, err := db.ReadDeviceByID(s.DeviceID)
	if err != nil {
		return err
	}
	u, err := db.ReadUserByID(dev.UserID)
	if err != nil {
		return err
	}

	r := permissions.GetUserRole(perm, u)

	if err = s.Validate(); err != nil {
		return err
	}
	s.Streamlimit = r.MaxStreams
	return db.Userdb.CreateStream(s)
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
	s, err := db.ReadStreamByID(streamID)
	if err != nil {
		return err
	}

	oldname := s.Name
	olddownlink := s.Downlink

	err = WriteObjectFromMap(s, updates)
	if err != nil {
		return err
	}

	if s.Name != oldname {
		return errors.New("ConnectorDB does not support modification of stream names")
	}

	// The stream schema is validated in users
	err = db.Userdb.UpdateStream(s)

	// If the stream is no longer downlink, delete the downlink substream
	if err == nil && olddownlink && !s.Downlink {
		db.DeleteStreamByID(streamID, "downlink")
	}
	return err
}

// DeleteStreamByID removes the stream
func (db *Database) DeleteStreamByID(streamID int64, substream string) error {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return err //Workaround #81
	}

	if substream != "" {
		//We just delete the substream
		err = db.DataStream.DeleteSubstream(strm.DeviceID, strm.StreamID, substream)
	} else {
		err = db.Userdb.DeleteStream(streamID)
		if err == nil {
			err = db.DataStream.DeleteStream(strm.DeviceID, strm.StreamID)
		}
	}
	return err
}
