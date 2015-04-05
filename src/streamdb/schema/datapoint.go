package schema

import (
	"fmt"
	"time"
)

//The Datapoint struct is used to encode a single element of information, ready to be Marshalled/unmarshalled into any of the various types
type Datapoint struct {
	Timestamp float64     `json:"t" xml:"t,attr"`
	Data      interface{} `json:"d"`
	Sender    string      `json:"o,omitempty" xml:"o,attr"`
	Stream    string      `json:"s,omitempty" xml:"s,attr"`

	schema *Schema //The schema allows to validate the data to make sure that it fits the accepted types
}

//DataBytes returns the byte array associated with the data of the datapoint
func (d *Datapoint) DataBytes() ([]byte, error) {
	return d.schema.Marshal(d.Data)
}

//IntTimestamp returns the unix nanoseconds timestamp
func (d *Datapoint) IntTimestamp() int64 {
	return int64(1e9 * d.Timestamp)
}

//String prints out a pretty string representation of the datapoint
func (d *Datapoint) String() string {
	s := fmt.Sprintf("[T=%s D=%v S=%s", time.Unix(0, d.IntTimestamp()), d.Data, d.Stream)
	if d.Sender != "" {
		s += " O=" + d.Sender
	}
	return s + "]"
}

//NewDatapoint reates a new uninitialized datapoint (which can be marshalled to)
func NewDatapoint(schema *Schema) *Datapoint {
	return &Datapoint{schema: schema}
}

//LoadDatapoint loads an existing data point into a datapoint struct
func LoadDatapoint(schema *Schema, timestamp int64, data []byte, sender string, stream string) (*Datapoint, error) {
	dp := NewDatapoint(schema)

	//Convert the data byte array to the wanted structure
	err := schema.Unmarshal(data, &dp.Data)

	//Convert the int nanosecond timestamp to a seconds timestamp
	dp.Timestamp = float64(timestamp) * 1e-9

	//Set up the sender and stream data (if applicable)
	dp.Sender = sender
	dp.Stream = stream

	return dp, err
}
