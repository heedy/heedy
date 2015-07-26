package messenger

import "connectordb/streamdb/datastream"

//Message is what is sent over NATS
type Message struct {
	Stream string                    `json:"stream" msgpack:"s,omitempty"`
	Data   datastream.DatapointArray `json:"data" msgpack:"d,omitempty"`
}
