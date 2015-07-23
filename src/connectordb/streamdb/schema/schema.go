package schema

/*
Package schema ensures that the given data conforms to the specified schema, and subsequently converts the given data to a byte array
suitable for storage in datastream
*/

import (
	"gopkg.in/vmihailenco/msgpack.v2"

	"connectordb/streamdb/util"

	"github.com/xeipuuv/gojsonschema"
)

//Schema is an object that given a JSONSchema can validate data and convert it to/from byte arrays
type Schema struct {
	jschema *gojsonschema.Schema
}

//IsValid checks for the datapoint's validity
func (s *Schema) IsValid(datapoint interface{}) bool {
	result, err := s.jschema.Validate(gojsonschema.NewGoLoader(datapoint))
	return (err == nil && result.Valid())
}

//Marshal a datapoint (assumed to be validity-checked) into a byte array ready to be written to the database
func (s *Schema) Marshal(datapoint interface{}) ([]byte, error) {
	return msgpack.Marshal(datapoint)
}

//Unmarshal the data into the given interface.
func (s *Schema) Unmarshal(data []byte, v interface{}) error {
	return util.MsgPackUnmarshal(data, v)
}

//NewSchema loads a schema from the JsonSchema string
func NewSchema(schema string) (*Schema, error) {
	sch, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema))
	return &Schema{sch}, err //sch might be nil, while Schema won't be, so it is critical that user checks for err rather than nil
}
