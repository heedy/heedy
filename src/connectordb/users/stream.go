package users

/** Package users provides an API for managing user information.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

import (
	"connectordb/datastream"
	"connectordb/schema"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"util"

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

func init() {
	// error triggered if we have size of 0, so don't set this to 0
	streamCache, _ = multicache.NewDefaultMulticache(10000)
}

type Stream struct {
	StreamId    int64  `modifiable:"nobody" json:"-"`
	Name        string `modifiable:"device" json:"name"`
	Nickname    string `modifiable:"device" json:"nickname,omitempty"`
	Description string `modifiable:"device" json:"description,omitempty"` // A public description
	Icon        string `modifiable:"device" json:"icon,omitempty"`        // A public icon in a data URI format, should be smallish 100x100?
	Type        string `modifiable:"root" json:"type"`
	DeviceId    int64  `modifiable:"nobody" json:"-"`
	Ephemeral   bool   `modifiable:"device" json:"ephemeral"`
	Downlink    bool   `modifiable:"device" json:"downlink"`
}

func (s *Stream) String() string {
	return fmt.Sprintf("[users.Stream | Id: %v, Name: %v, Nick: %v, Device: %v, Ephem: %v, Downlink: %v, Type: %v]",
		s.StreamId, s.Name, s.Nickname, s.DeviceId, s.Ephemeral, s.Downlink, s.Type)
}

// Checks if the fields are valid, e.g. we're not trying to change the name to blank.
func (s *Stream) ValidityCheck() error {

	_, err := s.GetSchema()
	if err != nil {
		return ErrInvalidSchema
	}

	if !IsValidName(s.Name) {
		return ErrInvalidUsername
	}

	return nil
}

func (d *Stream) RevertUneditableFields(originalValue Stream, p PermissionLevel) int {
	return revertUneditableFields(reflect.ValueOf(d), reflect.ValueOf(originalValue), p)
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

// Gets the jsonschema associated with this stream
func (s *Stream) GetSchema() (schema.Schema, error) {
	strmschema, ok := streamCache.Get(s.Type)
	if ok {
		return strmschema.(schema.Schema), nil
	}

	computedSchema, err := schema.NewSchema(s.Type)
	if err != nil || computedSchema == nil {
		return schema.Schema{}, err
	}

	streamCache.Add(s.Type, *computedSchema)
	return *computedSchema, nil
}

// CreateStream creates a new stream for a given device with the given name, schema and default values.
func (userdb *SqlUserDatabase) CreateStream(Name, Type string, DeviceId int64) error {

	if !IsValidName(Name) {
		return InvalidNameError
	}

	// Validate that the schema is correct
	if _, err := schema.NewSchema(Type); err != nil {
		return ErrInvalidSchema
	}

	// Validate no object subtypes (they are valid, but not in this database
	// due to ml considerations)
	if util.SchemaContainsObjectFields(Type) {
		return ErrInvalidSchema
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

	if err == sql.ErrNoRows {
		return nil, ErrStreamNotFound
	}

	return &stream, err
}

// ReadStreamById fetches the stream with the given id and returns it, or nil if
// no such stream exists.
func (userdb *SqlUserDatabase) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	var stream Stream

	err := userdb.Get(&stream, "SELECT * FROM Streams WHERE DeviceId = ? AND Name = ? LIMIT 1;", DeviceId, streamName)

	if err == sql.ErrNoRows {
		return nil, ErrStreamNotFound
	}

	return &stream, err
}

func (userdb *SqlUserDatabase) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	var streams []Stream

	err := userdb.Select(&streams, "SELECT * FROM Streams WHERE DeviceId = ?;", DeviceId)

	if err == sql.ErrNoRows {
		return nil, ErrStreamNotFound
	}

	return streams, err
}

func (userdb *SqlUserDatabase) ReadStreamsByUser(UserId int64) ([]Stream, error) {
	var streams []Stream

	err := userdb.Select(&streams, `SELECT s.* FROM Streams s, Devices d, Users u
	WHERE
		u.UserId = ? AND
		d.UserId = u.UserId AND
		s.DeviceId = d.DeviceId`, UserId)

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

	_, err := userdb.Exec(`UPDATE Streams SET
	    Name = ?,
		Nickname = ?,
		Description = ?,
		Icon = ?,
	    Type = ?,
	    DeviceId = ?,
	    Ephemeral = ?,
	    Downlink = ?
	    WHERE StreamId = ?;`,
		stream.Name,
		stream.Nickname,
		stream.Description,
		stream.Icon,
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
