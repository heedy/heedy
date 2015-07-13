package authoperator

import (
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"

	"github.com/nats-io/nats"
)

//SubscribeUserByID subscribes to everything a user creates
func (o *AuthOperator) SubscribeUserByID(userID int64, chn chan operator.Message) (*nats.Subscription, error) {
	usr, err := o.Db.ReadUserByID(userID)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToUser(usr).Gte(users.USER) {
		return o.Db.SubscribeUserByID(userID, chn)
	}
	return nil, ErrPermissions
}

//SubscribeDeviceByID subscribes to all streams of the given device
func (o *AuthOperator) SubscribeDeviceByID(deviceID int64, chn chan operator.Message) (*nats.Subscription, error) {
	readdev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	if dev.RelationToDevice(readdev).Gte(users.DEVICE) {
		return o.Db.SubscribeDeviceByID(deviceID, chn)
	}
	return nil, ErrPermissions
}

//SubscribeStreamByID subscribes to the given stream by ID
func (o *AuthOperator) SubscribeStreamByID(streamID int64, substream string, chn chan operator.Message) (*nats.Subscription, error) {
	dev, err := o.Device()
	if err != nil {
		return nil, err
	}
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	sdevice, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return nil, err
	}

	if dev.RelationToStream(&strm.Stream, sdevice).Gte(users.DEVICE) {
		return o.Db.SubscribeStreamByID(streamID, substream, chn)
	}
	return nil, ErrPermissions
}
