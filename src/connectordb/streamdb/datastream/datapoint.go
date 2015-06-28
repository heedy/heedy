package datastream

import (
	"fmt"
	"reflect"

	"github.com/xeipuuv/gojsonschema"
)

//Datapoint is the
type Datapoint struct {
	// Timestamp is the Unix timestamp as a float
	Timestamp float64 `json:"t,omitempty" msgpack:"t,omitempty"`
	// The actual data associated with this point
	Data interface{} `json:"d,omitempty" msgpack:"d,omitempty"`
	// Sender is optional, path to a device.
	Sender string `json:"o,omitempty" msgpack:"o,omitempty"`
}

//String prints out a pretty string representation of the datapoint
func (d *Datapoint) String() string {
	s := fmt.Sprintf("[T=%.3f D=%v", d.Timestamp, d.Data)
	if d.Sender != "" {
		s += " S=" + d.Sender
	}
	return s + "]"
}

//IsEqual checks if the datapoint is equal to another datapoint
func (d *Datapoint) IsEqual(dp Datapoint) bool {
	return (dp.Timestamp == d.Timestamp && dp.Sender == d.Sender && reflect.DeepEqual(d.Data, dp.Data))
}

//HasSchema returns true if the datapoint conforms to the passed schema
func (d *Datapoint) HasSchema(schema *gojsonschema.Schema) bool {
	res, err := schema.Validate(gojsonschema.NewGoLoader(d.Data))
	return err == nil && res.Valid()
}
