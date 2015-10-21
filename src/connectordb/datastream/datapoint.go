package datastream

import (
	"util"
	"fmt"
	"reflect"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/vmihailenco/msgpack.v2"
)

//Datapoint is the
type Datapoint struct {
	// Timestamp is the Unix timestamp as a float
	Timestamp float64 `json:"t,omitempty" msgpack:"t,omitempty" duck:"t"`
	// The actual data associated with this point
	Data interface{} `json:"d,omitempty" msgpack:"d,omitempty" duck:"d"`
	// Sender is optional, path to a device.
	Sender string `json:"o,omitempty" msgpack:"o,omitempty" duck:"o"`
}

//DatapointFromBytes reads a datapoint from its byte representation
func DatapointFromBytes(data []byte) (d Datapoint, err error) {
	//We need msgpack to be unmarshalled with string maps, rather than interface maps.
	err = util.MsgPackUnmarshal(data, &d)

	return d, err
}

//String prints out a pretty string representation of the datapoint
func (d *Datapoint) String() string {
	if d == nil {
		return "nil"
	}
	s := fmt.Sprintf("[T=%.3f D=%v", d.Timestamp, d.Data)
	if d.Sender != "" {
		s += " S=" + d.Sender
	}
	return s + "]"
}

//Bytes returns the msgpack marshalled representation of the datapoint
func (d *Datapoint) Bytes() ([]byte, error) {
	return msgpack.Marshal(d)
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

// Copy creates a shallow copy of the datapoint
func (dp *Datapoint) Copy() *Datapoint {
	var result Datapoint
	result.Timestamp = dp.Timestamp
	result.Data = dp.Data //Note: most likely this is not a deep copy
	result.Sender = dp.Sender
	return &result
}

//NewDatapoint returns a datapoint with the current timestamp
func NewDatapoint() Datapoint {
	var dp Datapoint
	dp.Timestamp = float64(time.Now().UnixNano()) * 1e-9
	return dp
}
