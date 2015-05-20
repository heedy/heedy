package streamdb

/*
Package Messenger is a package that implements the pub/sub messaging system used for streaming uplinks and downlinks
as well as the messaging system that allows real-time low-latency data analysis.
*/

import (
	"connectordb/streamdb/operator"
	"strings"

	"github.com/apcera/nats"
)

//Package messenger provides a simple messaging service using gnatsd, which can be used to
//send fast messages to a given user/device/stream from a given user/device

//Messenger holds an open connection to the gnatsd daemon
type Messenger struct {
	Conn  *nats.Conn        //The NATS connection
	Econn *nats.EncodedConn //The Encoded conn, ie, a data message
}

//Close shuts down a Messenger
func (m *Messenger) Close() {
	m.Econn.Close()
	m.Conn.Close()
}

//ConnectMessenger initializes a connection with the gnatsd messenger. Allows daisy-chaining errors
func ConnectMessenger(url string, err error) (*Messenger, error) {
	if err != nil {
		return nil, err
	}

	conn, err := nats.Connect("nats://" + url)
	if err != nil {
		return nil, err
	}
	econn, err := nats.NewEncodedConn(conn, "gob")
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Messenger{conn, econn}, nil
}

//Publish sends the given message over the connection
func (m *Messenger) Publish(routing string, msg operator.Message) error {
	routing = strings.Replace(routing, "/", ".", -1)
	if routing[len(routing)-1] == '.' {
		routing = routing[0 : len(routing)-1]
	}
	return m.Econn.Publish(routing, msg)
}

//Subscribe creates a subscription for the given routing string. The routing string is of the format:
//  [user]/[device]/[stream]/[substream//]
//In order to skip something, you can use wildcards, and to skip "the rest" you can use ">" (this is literally the gnatsd routing)
//An example of subscribing to all posts by sender user user1:
//  msgr.Subscribe("user1/>",chn)
//An example of subscribing to everything is:
//	msgr.Subscribe(">",chn)
//Subscribing to a stream is:
// msgr.Subscribe("user/device/stream")
func (m *Messenger) Subscribe(routing string, chn chan operator.Message) (*nats.Subscription, error) {
	return m.Econn.BindRecvChan(strings.Replace(routing, "/", ".", -1), chn)
}

//Flush makes sure all commands are acknowledged by the server
func (m *Messenger) Flush() {
	m.Econn.Flush()
}
