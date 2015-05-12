package streamdb

import (
	"encoding/json"
	"streamdb/schema"
	"streamdb/users"
)

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

	err = json.Unmarshal([]byte(s.Type), &schemamap)

	return Stream{*s, schemamap, strmschema}, err
}
