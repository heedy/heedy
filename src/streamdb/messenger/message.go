package messenger

import (
	"strings"
)

//The Message is a struct holding field data which is sent through Messenger
type Message struct {
	To     string //The To field is the stream that the message is aimed at
	From   string //The from field is the device sending the message
	Prefix string //The Prefix is a special "message type" identifier.
	Dtype  string //The typestring of the data in the message
	Data   []byte //The data byte array holding a timebatchDB DatapointArray
}

//String returns a stringified representation of the message
func (m Message) String() string {
	return "[To=" + m.To + " From=" + m.From + " Pre=" + m.Prefix + "]"
}

//IsFromSelf returns whether or not the given message is sent from its own device.
func (m Message) IsFromSelf() bool {
	return strings.HasPrefix(m.To, m.From)
}
