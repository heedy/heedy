package users

/** Package users provides an API for managing user information.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

import "reflect"

type Stream struct {
	StreamId  int64  `modifiable:"nobody" json:"-"`
	Name      string `modifiable:"device" json:"name"`
	Nickname  string `modifiable:"device" json:"nickname,omitempty"`
	Type      string `modifiable:"root" json:"-"`
	DeviceId  int64  `modifiable:"nobody" json:"-"`
	Ephemeral bool   `modifiable:"device" json:"ephemeral,omitempty"`
	Downlink  bool   `modifiable:"device" json:"downlink,omitempty"`
}

// Checks if the fields are valid, e.g. we're not trying to change the name to blank.
func (s *Stream) ValidityCheck() error {
	if !IsValidName(s.Name) {
		return InvalidUsernameError
	}

	return nil
}

func (d *Stream) RevertUneditableFields(originalValue Stream, p PermissionLevel) int {
	return revertUneditableFields(reflect.ValueOf(d), reflect.ValueOf(originalValue), p)
}

// CreateStream creates a new stream for a given device with the given name, schema and default values.
func (userdb *SqlUserDatabase) CreateStream(Name, Type string, DeviceId int64) error {

	if !IsValidName(Name) {
		return InvalidNameError
	}

	_, err := userdb.Exec(`INSERT INTO Streams
	    (	Name,
	        Type,
	        DeviceId) VALUES (?,?,?);`, Name, Type, DeviceId)

	return err
}

// ReadStreamById fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *SqlUserDatabase) ReadStreamById(StreamId int64) (*Stream, error) {
	var stream Stream

	err := userdb.Get(&stream, "SELECT * FROM Streams WHERE StreamId = ? LIMIT 1;", StreamId)

	return &stream, err
}

// ReadStreamById fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *SqlUserDatabase) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	var stream Stream

	err := userdb.Get(&stream, "SELECT * FROM Streams WHERE DeviceId = ? AND Name = ? LIMIT 1;", DeviceId, streamName)

	return &stream, err
}

func (userdb *SqlUserDatabase) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	var streams []Stream

	err := userdb.Select(&streams, "SELECT * FROM Streams WHERE DeviceId = ?;", DeviceId)

	return streams, err
}

// UpdateStream updates the stream with the given ID with the provided data
// replacing all prior contents.
func (userdb *SqlUserDatabase) UpdateStream(stream *Stream) error {
	if stream == nil {
		return InvalidPointerError
	}

	if err := stream.ValidityCheck(); err != nil {
		return err
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
func (userdb *SqlUserDatabase) DeleteStream(Id int64) error {
	result, err := userdb.Exec(`DELETE FROM Streams WHERE StreamId = ?;`, Id)
	return getDeleteError(result, err)
}
