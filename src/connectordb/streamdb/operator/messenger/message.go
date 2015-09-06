package messenger

import "connectordb/streamdb/datastream"

//Message is what is sent over NATS
type Message struct {
	Stream    string                    `json:"stream" msgpack:"s,omitempty"`
	Transform string                    `json:"transform,omitempty" msgpack:"t,omitempty"`
	Data      datastream.DatapointArray `json:"data" msgpack:"d,omitempty"`
}
