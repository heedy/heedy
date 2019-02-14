/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package messenger

import "connectordb/datastream"

//Message is what is sent over NATS
type Message struct {
	Stream    string                    `json:"stream" msgpack:"s,omitempty"`
	Transform string                    `json:"transform,omitempty" msgpack:"t,omitempty"`
	Data      datastream.DatapointArray `json:"data" msgpack:"d,omitempty"`
}
