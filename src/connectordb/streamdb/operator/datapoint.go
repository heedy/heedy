package operator

import (
	"connectordb/streamdb/schema"
	"fmt"
	"time"
)

//The Datapoint struct is used to encode a single element of information, ready
// to be Marshalled/unmarshalled into any of the various types
type Datapoint struct {
	// Unix timestamp as a float
	Timestamp float64 `json:"t" xml:"t,attr"`
	// The actual data associated with this point
	Data   interface{} `json:"d"`
	// Sender is optional, path to a device.
	Sender string      `json:"o,omitempty" xml:"o,attr"`
	// Stream may not be used yet.
	Stream string      `json:"s,omitempty" xml:"s,attr"`
}

// Creates a new datapoint with empty sender, stream and the current timestamp
// good for creating and parsing data on the fly
func NewDatapoint(Data interface{}) Datapoint {
	var dp Datapoint
	dp.Data = Data
	dp.Timestamp = float64(time.Now().UnixNano()) * 1e-9
	return dp
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

//LoadDatapoint loads an existing data point into a datapoint struct
func LoadDatapoint(schema *schema.Schema, timestamp int64, data []byte, sender string, stream string, err error) (*Datapoint, error) {
	if err != nil {
		return nil, err
	}

	var dp Datapoint

	//Convert the data byte array to the wanted structure
	err = schema.Unmarshal(data, &dp.Data)

	//Convert the int nanosecond timestamp to a seconds timestamp
	dp.Timestamp = float64(timestamp) * 1e-9

	//Set up the sender and stream data (if applicable)
	dp.Sender = sender
	dp.Stream = stream

	return &dp, err
}

//The Message is a struct holding field data which is sent through Messenger
type Message struct {
	Stream string      `json:"stream"` //The stream that the message is aimed at
	Data   []Datapoint `json:"data"`   //The datapoints associated with the message
}

//String returns a stringified representation of the message
func (m Message) String() string {
	return "[S=" + m.Stream + "]"
}
