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
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/json"
)

var (
	ErrSchema        = errors.New("The datapoints did not match the stream's schema")
	ErrInvalidSchema = errors.New("The provided schema is not a valid JSONSchema")
	schemaCache      *multicache.Multicache
)

type Stream struct {
	StreamID    int64  `json:"-" permissions:"-"`
	Name        string `json:"name" permissions:"name"`
	Nickname    string `json:"nickname" permissions:"nickname"`
	Description string `json:"description" permissions:"description"` // A public description
	Icon        string `json:"icon" permissions:"icon"`               // A public icon in a data URI format, should be smallish 100x100?
	Schema      string `json:"schema" permissions:"schema"`
	Datatype    string `json:"datatype" permissions:"datatype"`
	DeviceID    int64  `json:"-" permissions:"-"`
	Ephemeral   bool   `json:"ephemeral" permissions:"ephemeral"`
	Downlink    bool   `json:"downlink" permissions:"downlink"`
}

// The struct passed in to create a stream
type StreamMaker struct {
	Stream

	Streamlimit int64 `json:"-"`
}

// minifyAndValidateSchema minifies the schema, and makes sure that it is valid.
// The function also adds the schema to the cache if it is not there already.
func minifyAndValidateSchema(s string) (string, error) {
	if s == "" {
		s = "{}"
	}

	m := minify.New()
	m.AddFunc("text/json", json.Minify)

	minified, err := m.String("text/json", s)
	if err != nil {
		return "", err
	}

	// Now check the validity of the schema - if it is in the cache, it must be valid.
	// If not, make sure it is valid, and cache the result
	_, ok := schemaCache.Get(minified)
	if ok {
		return minified, nil
	}

	computedSchema, err := schema.NewSchema(minified)
	if err != nil {
		return minified, err
	}
	schemaCache.Add(minified, *computedSchema)

	return minified, nil
}

// Validate ensures that the maker holds allowed values
func (s *StreamMaker) Validate() error {
	return s.ValidityCheck()
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
	if s.Schema == "" {
		s.Schema = "{}"
	}
	strmschema, ok := schemaCache.Get(s.Schema)
	if ok {
		return strmschema.(schema.Schema), nil
	}

	computedSchema, err := schema.NewSchema(s.Schema)
	if err != nil || computedSchema == nil {
		return schema.Schema{}, err
	}

	schemaCache.Add(s.Schema, *computedSchema)
	return *computedSchema, nil
}

// CreateStream creates a new stream for a given device with the given name, schema and default values
// It is assumed that streammaker.Validate() has already been run on the stream
func (userdb *SqlUserDatabase) CreateStream(s *StreamMaker) error {

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

	// Validate that the schema is correct
	minSchema, err := minifyAndValidateSchema(s.Schema)
	if err != nil {
		return err
	}

	_, err = userdb.Exec(`INSERT INTO streams
		(	name,
			schema,
			deviceid,
			description,
			datatype,
			icon,
			nickname,
			ephemeral,
			downlink) VALUES (?,?,?,?,?,?,?,?,?);`, s.Name, minSchema, s.DeviceID,
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

// UpdateStream updates the stream with the given ID with the provided data
// replacing all prior contents.
func (userdb *SqlUserDatabase) UpdateStream(stream *Stream) error {
	if stream == nil {
		return InvalidPointerError
	}

	// Validate that the schema is correct
	minSchema, err := minifyAndValidateSchema(stream.Schema)
	if err != nil {
		return err
	}

	_, err = userdb.Exec(`UPDATE streams SET
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
		minSchema,
		stream.Datatype,
		stream.DeviceID,
		stream.Ephemeral,
		stream.Downlink,
		stream.StreamID)

	return err
}

// DeleteStream removes a stream from the database
func (userdb *SqlUserDatabase) DeleteStream(ID int64) error {
	result, err := userdb.Exec(`DELETE FROM streams WHERE streamid = ?;`, ID)
	return getDeleteError(result, err)
}

// DevStream includes the device name for use when querying directly from a user
type DevStream struct {
	Stream
	Device string `json:"device" permissions:"device"`
}

// ReadStreamsByUser returns a user's streams along with parent device name.
// If downlink is true, returns only streams that are downlinks
// If public is true, only returns streams of public devices.
// If hidehidden is true, only returns the streams that belong to visible devices
func (userdb *SqlUserDatabase) ReadStreamsByUser(UserID int64, public, downlink, hidehidden bool) ([]*DevStream, error) {
	var streams []*DevStream

	query := `SELECT s.*, d.name AS device FROM streams s
		INNER JOIN devices d ON s.deviceid = d.deviceid
		WHERE d.userid = ?`

	// sqlite has issues with TRUE/FALSE, so we let sqlx take care of converting
	// things into the correct variable types: we send booleans in as parameters
	params := []interface{}{UserID}

	if downlink {
		query += " AND s.downlink = ?"
		params = append(params, true)
	}

	if public {
		query += " AND d.public = ?"
		params = append(params, true)
	}

	if hidehidden {
		query += " AND d.isvisible = ?"
		params = append(params, true)
	}

	err := userdb.Select(&streams, query, params...)

	if err == sql.ErrNoRows {
		err = nil
	}

	return streams, err
}
