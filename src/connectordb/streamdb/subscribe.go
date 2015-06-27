package streamdb

import (
	"connectordb/streamdb/operator"

	"github.com/nats-io/nats"
)

//SubscribeUserByID subscribes to everything a user creates
func (o *Database) SubscribeUserByID(userID int64, chn chan operator.Message) (*nats.Subscription, error) {
	return o.msg.Subscribe(getTimebatchUserName(userID)+"/*/*", chn)
}

//SubscribeDeviceByID subscribes to all streams of the given device
func (o *Database) SubscribeDeviceByID(deviceID int64, chn chan operator.Message) (*nats.Subscription, error) {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	return o.msg.Subscribe(getTimebatchDeviceName(dev)+"/*", chn)
}

//SubscribeStreamByID subscribes to the given stream by ID
func (o *Database) SubscribeStreamByID(streamID int64, substream string, chn chan operator.Message) (*nats.Subscription, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	routing, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return nil, err
	}
	return o.msg.Subscribe(routing+substream, chn)
}
