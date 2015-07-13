package operator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/schema"
	"connectordb/streamdb/users"
	"encoding/json"
	"errors"

	"github.com/josephlewis42/multicache"
)

const (
	schemaCacheSize = 1000
)

var (
	//ErrSchema is thrown when schemas don't match
	ErrSchema   = errors.New("The datapoints did not match the stream's schema")
	schemaCache *multicache.Multicache
)

func init() {
	schemaCache, _ = multicache.NewDefaultMulticache(schemaCacheSize)
}

//Stream is a wrapper for the users.Stream object which encodes the schema and other parts of a stream
type Stream struct {
	users.Stream
	Schema map[string]interface{} `json:"schema"` //This allows the JsonSchema to be directly unmarshalled

	//These are used internally for the stream to work out
	s *schema.Schema //The schema associated with the stream
}

//NewStream returns a new stream object
func NewStream(s *users.Stream, err error) (Stream, error) {
	if err != nil {
		return Stream{}, err
	}

	strmschema, err := schema.NewSchema(s.Type)
	if err != nil {
		return Stream{}, err
	}

	var schemamap map[string]interface{}
	var sm interface{}
	ok := false

	// We initialized it to empty, go on without the cache.
	if schemaCache != nil {
		sm, ok = schemaCache.Get(s.Type)
	}

	if ok {
		schemamap = sm.(map[string]interface{})
	} else {
		err = json.Unmarshal([]byte(s.Type), &schemamap)
	}

	return Stream{*s, schemamap, strmschema}, err
}

//Validate ensures the array of datapoints conforms to the schema and such
func (s *Stream) Validate(data datastream.DatapointArray) bool {
	for i := range data {
		if !s.s.IsValid(data[i].Data) {
			return false
		}
	}
	return true
}

func (s *Stream) GetSchema() *schema.Schema {
	return s.s
}
