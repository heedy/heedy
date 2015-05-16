package streamdb

import (
	"fmt"
	"streamdb/schema"
	"time"
)

//The Datapoint struct is used to encode a single element of information, ready to be Marshalled/unmarshalled into any of the various types
type Datapoint struct {
	Timestamp float64     `json:"t" xml:"t,attr"`
	Data      interface{} `json:"d"`
	Sender    string      `json:"o,omitempty" xml:"o,attr"`
	Stream    string      `json:"s,omitempty" xml:"s,attr"`
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
	To     string      //The To field is the stream that the message is aimed at
	From   string      //The from field is the device sending the message
	Prefix string      //The Prefix is a special "message type" identifier.
	Data   []Datapoint //The datapoints associated with the message
}

//String returns a stringified representation of the message
func (m Message) String() string {
	return "[To=" + m.To + " From=" + m.From + " Pre=" + m.Prefix + "]"
}
