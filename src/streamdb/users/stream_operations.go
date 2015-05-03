package users

/** Package users provides an API for managing user information.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

import "reflect"


type Stream struct {
	StreamId  int64  `modifiable:"nobody"`
	Name      string `modifiable:"nobody"`
	Nickname  string `modifiable:"user"`
	Type      string `modifiable:"root"`
	DeviceId  int64  `modifiable:"nobody"`
	Ephemeral bool   `modifiable:"user"`
	Downlink  bool   `modifiable:"user"`
}


func (d *Stream) RevertUneditableFields(originalValue Stream, p PermissionLevel) {
	revertUneditableFields(reflect.ValueOf(*d), reflect.ValueOf(originalValue), p)
}



// CreateStream creates a new stream for a given device with the given name, schema and default values.
func (userdb *UserDatabase) CreateStream(Name, Type string, DeviceId int64) error {
	_, err := userdb.Exec(`INSERT INTO Streams
	    (	Name,
	        Type,
	        DeviceId) VALUES (?,?,?);`, Name, Type, DeviceId)

	return err
}

// ReadStreamById fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *UserDatabase) ReadStreamById(StreamId int64) (*Stream, error) {
	var stream Stream

	err := userdb.Get(&stream, "SELECT * FROM Streams WHERE StreamId = ? LIMIT 1;", StreamId)

	return &stream, err
}

// ReadStreamById fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *UserDatabase) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	var stream Stream

	err := userdb.Get(&stream, "SELECT * FROM Streams WHERE DeviceId = ? AND Name = ? LIMIT 1;", DeviceId, streamName)

	return &stream, err
}

func (userdb *UserDatabase) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	var streams []Stream

	err := userdb.Select(&streams, "SELECT * FROM Streams WHERE DeviceId = ?;", DeviceId)

	return streams, err
}

// UpdateStream updates the stream with the given ID with the provided data
// replacing all prior contents.
func (userdb *UserDatabase) UpdateStream(stream *Stream) error {
	if stream == nil {
		return ERR_INVALID_PTR
	}

	_, err := userdb.Exec(`UPDATE Streams SET
	    Name = ?,
		Nickname = ?,
	    Type = ?,
	    DeviceId = ?,
	    Ephemeral = ?,
	    Downlink = ?
	    WHERE StreamId = ?;`,
		stream.Name,
		stream.Nickname,
		stream.Type,
		stream.DeviceId,
		stream.Ephemeral,
		stream.Downlink,
		stream.StreamId)

	return err
}

// DeleteStream removes a stream from the database
func (userdb *UserDatabase) DeleteStream(Id int64) error {
	_, err := userdb.Exec(`DELETE FROM Streams WHERE StreamId = ?;`, Id)
	return err
}
