/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package users

import (
	"connectordb/datastream"
	"connectordb/schema"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/josephlewis42/multicache"
)

const (
	schemaCacheSize = 1000
)

var (
	ErrSchema        = errors.New("The datapoints did not match the stream's schema")
	ErrInvalidSchema = errors.New("The provided schema is not a valid JSONSchema")
	streamCache      *multicache.Multicache
)

type Stream struct {
	StreamID    int64  `json:"-"`
	Name        string `json:"name"`
	Nickname    string `json:"nickname"`
	Description string `json:"description"` // A public description
	Icon        string `json:"icon"`        // A public icon in a data URI format, should be smallish 100x100?
	Schema      string `json:"schema"`
	Datatype    string `json:"datatype"`
	DeviceID    int64  `json:"-"`
	Ephemeral   bool   `json:"ephemeral"`
	Downlink    bool   `json:"downlink"`
}

// The struct passed in to create a stream
type StreamMaker struct {
	Stream

	Streamlimit int64 `json:"-"`
}

// Validate ensures that the maker holds allowed values
func (s *StreamMaker) Validate() error {
	return s.ValidityCheck()
}

func (s *Stream) String() string {
	return fmt.Sprintf("[users.Stream | Id: %v, Name: %v, Nick: %v, Device: %v, Ephem: %v, Downlink: %v, Schema: %v]",
		s.StreamID, s.Name, s.Nickname, s.DeviceID, s.Ephemeral, s.Downlink, s.Schema)
}

// ValidityCheck checks if the fields are valid, e.g. we're not trying to change the name to blank.
func (s *Stream) ValidityCheck() error {

	_, err := s.GetSchema()
	if err != nil {
		return ErrInvalidSchema
	}

	if !IsValidName(s.Name) {
		return ErrInvalidUsername
	}
	err = validateIcon(s.Icon)
	return err
}

// Validate ensures the array of datapoints conforms to the schema and such
func (s *Stream) Validate(data datastream.DatapointArray) bool {
	schema, err := s.GetSchema()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		return false
	}

	for _, datum := range data {
		if !schema.IsValid(datum.Data) {
			return false
		}
	}

	return true
}

// GetSchema gets the jsonschema associated with this stream
func (s *Stream) GetSchema() (schema.Schema, error) {
	strmschema, ok := streamCache.Get(s.Schema)
	if ok {
		return strmschema.(schema.Schema), nil
	}

	computedSchema, err := schema.NewSchema(s.Schema)
	if err != nil || computedSchema == nil {
		return schema.Schema{}, err
	}

	streamCache.Add(s.Schema, *computedSchema)
	return *computedSchema, nil
}

// CreateStream creates a new stream for a given device with the given name, schema and default values
// It is assumed that streammaker.Validate() has already been run on the stream
func (userdb *SqlUserDatabase) CreateStream(s *StreamMaker) error {

	// Validate that the schema is correct
	if _, err := schema.NewSchema(s.Schema); err != nil {
		return ErrInvalidSchema
	}

	if s.Streamlimit > 0 {
		// TODO: This should be done in an SQL transaction due to possible timing bugs
		num, err := userdb.CountStreamsForDevice(s.DeviceID)
		if err != nil {
			return err
		}
		if num >= s.Streamlimit {
			return errors.New("Cannot create stream: Exceeded maximum stream number for device.")
		}
	}

	_, err := userdb.Exec(`INSERT INTO streams
		(	name,
			schema,
			deviceid,
			description,
			datatype,
			icon,
			nickname,
			ephemeral,
			downlink) VALUES (?,?,?,?,?,?,?,?,?);`, s.Name, s.Schema, s.DeviceID,
		s.Description, s.Datatype, s.Icon, s.Nickname, s.Ephemeral, s.Downlink)

	if err != nil && strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint ") {
		return errors.New("Stream with this name already exists")
	}
	return err
}

// ReadStreamByID fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *SqlUserDatabase) ReadStreamByID(StreamID int64) (*Stream, error) {
	var stream Stream

	err := userdb.Get(&stream, "SELECT * FROM streams WHERE streamid = ? LIMIT 1;", StreamID)

	if err == sql.ErrNoRows {
		return nil, ErrStreamNotFound
	}

	return &stream, err
}

// ReadStreamByDeviceIDAndName fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *SqlUserDatabase) ReadStreamByDeviceIDAndName(DeviceID int64, streamName string) (*Stream, error) {
	var stream Stream

	err := userdb.Get(&stream, "SELECT * FROM streams WHERE deviceid = ? AND name = ? LIMIT 1;", DeviceID, streamName)

	if err == sql.ErrNoRows {
		return nil, ErrStreamNotFound
	}

	return &stream, err
}

func (userdb *SqlUserDatabase) ReadStreamsByDevice(DeviceID int64) ([]*Stream, error) {
	var streams []*Stream

	err := userdb.Select(&streams, "SELECT * FROM streams WHERE deviceid = ?;", DeviceID)

	if err == sql.ErrNoRows {
		return nil, ErrStreamNotFound
	}

	return streams, err
}

func (userdb *SqlUserDatabase) ReadStreamsByUser(UserID int64) ([]*Stream, error) {
	var streams []*Stream

	err := userdb.Select(&streams, `SELECT s.* FROM streams s, devices d, users u
	WHERE
		u.userid = ? AND
		d.userid = u.userid AND
		s.deviceid = d.deviceid`, UserID)

	if err == sql.ErrNoRows {
		return nil, ErrStreamNotFound
	}

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

	_, err := userdb.Exec(`UPDATE streams SET
		name = ?,
		nickname = ?,
		description = ?,
		icon = ?,
		schema = ?,
		datatype= ?,
		deviceid = ?,
		ephemeral = ?,
		downlink = ?
		WHERE streamid= ?;`,
		stream.Name,
		stream.Nickname,
		stream.Description,
		stream.Icon,
		stream.Schema,
		stream.Datatype,
		stream.DeviceID,
		stream.Ephemeral,
		stream.Downlink,
		stream.StreamID)

	return err
}

// DeleteStream removes a stream from the database
func (userdb *SqlUserDatabase) DeleteStream(Id int64) error {
	result, err := userdb.Exec(`DELETE FROM streams WHERE streamid = ?;`, Id)
	return getDeleteError(result, err)
}
