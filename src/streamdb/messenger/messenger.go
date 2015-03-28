package messenger

/*
Package Messenger is a package that implements the pub/sub messaging system used for streaming uplinks and downlinks
as well as the messaging system that allows real-time low-latency data analysis.
*/

import (
	"strings"

	"github.com/apcera/nats"
)

//Package messenger provides a simple messaging service using gnatsd, which can be used to
//send fast messages to a given user/device/stream from a given user/device

//Messenger holds an open connection to the gnatsd daemon
type Messenger struct {
	conn  *nats.Conn        //The NATS connection
	econn *nats.EncodedConn //The Encoded conn, ie, a data message
}

//Close shuts down a Messenger
func (m *Messenger) Close() {
	m.econn.Close()
	m.conn.Close()
}

//Connect initializes a connection with the gnatsd messenger.
func Connect(url string) (*Messenger, error) {
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
func (m *Messenger) Publish(msg Message) error {
	routing := msg.Prefix + "/" + msg.To + "/" + msg.From
	routing = strings.Replace(routing, "/", ".", -1)
	return m.econn.Publish(routing, msg)
}

//Subscribe creates a subscription for the given routing string. The routing string is of the format:
//  [prefix]/[user]/[device]/[stream]/[senderuser]/[senderdevice]
//In order to skip something, you can use wildcards, and to skip "the rest" you can use ">" (this is literally the gnatsd routing)
//An example of subscribing to all posts by sender user user1:
//  msgr.Subscribe("*/*/*/*/user1/>",chn)
//An example of subscribing to everything is:
//	msgr.Subscribe(">",chn)
func (m *Messenger) Subscribe(routing string, chn chan Message) (*nats.Subscription, error) {
	return m.econn.BindRecvChan(strings.Replace(routing, "/", ".", -1), chn)
}

//SubscribeSenderDevice subscribes to all messages sent by the given device. Since the device can send messages that are not yet acknowledged by their
//respective receiving devices (downlinks), can subscribe by prefix also. To ignore prefix, just use * for the prefix string
func (m *Messenger) SubscribeSenderDevice(prefix string, sender string, chn chan Message) (*nats.Subscription, error) {
	return m.Subscribe(prefix+"/*/*/*/"+sender, chn)
}

//SubscribeReceiverDevice subscribes to all messages sent TO a given device. This includes the device writing its own uplink streams.
//Since certain messages can be acknowledged or sent multiple times, can subscribe by prefix. Use * to ignore prefix.
func (m *Messenger) SubscribeReceiverDevice(prefix string, receiver string, chn chan Message) (*nats.Subscription, error) {
	return m.Subscribe(prefix+"/"+receiver+"/>", chn)
}

//SubscribeStream subscribes to all messages sent to a given stream. Since messages can be sent to downlink by other devices before acknowledgement,
//allows to filter by the given prefix. Use the wildcard * to ignore prefix.
func (m *Messenger) SubscribeStream(prefix string, stream string, chn chan Message) (*nats.Subscription, error) {
	return m.Subscribe(prefix+"/"+stream+"/>", chn)
}

//SubscribePrefix allwos to subscribe to all messages with the given prefix.
func (m *Messenger) SubscribePrefix(prefix string, chn chan Message) (*nats.Subscription, error) {
	return m.Subscribe(prefix+"/>", chn)
}

//Flush makes sure all commands are acknowledged by the server
func (m *Messenger) Flush() {
	m.econn.Flush()
}
