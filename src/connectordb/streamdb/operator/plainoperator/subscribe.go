package plainoperator

import (
	"connectordb/streamdb/operator/messenger"

	"github.com/nats-io/nats"
)

//SubscribeUserByID subscribes to everything a user creates
func (o *PlainOperator) SubscribeUserByID(userID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	usr, err := o.ReadUserByID(userID)
	if err != nil {
		return nil, err
	}
	return o.msg.Subscribe(usr.Name+"/*/*", chn)
}

//SubscribeDeviceByID subscribes to all streams of the given device
func (o *PlainOperator) SubscribeDeviceByID(deviceID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	usr, err := o.ReadUserByID(dev.UserId)
	if err != nil {
		return nil, err
	}
	return o.msg.Subscribe(usr.Name+"/"+dev.Name+"/*", chn)
}

//SubscribeStreamByID subscribes to the given stream by ID
func (o *PlainOperator) SubscribeStreamByID(streamID int64, substream string, chn chan messenger.Message) (*nats.Subscription, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	routing, err := o.getStreamPath(strm)
	if err != nil {
		return nil, err
	}
	if substream != "" {
		routing = routing + "/" + substream
	}
	return o.msg.Subscribe(routing, chn)
}
