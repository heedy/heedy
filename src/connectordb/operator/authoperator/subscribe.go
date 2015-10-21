package authoperator

import (
	"connectordb/operator/messenger"
	"connectordb/users"

	"github.com/nats-io/nats"
)

//SubscribeUserByID subscribes to everything a user creates
func (o *AuthOperator) SubscribeUserByID(userID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	usr, err := o.BaseOperator.ReadUserByID(userID)
	if err != nil {
		return nil, err
	}

	if _, err := o.permissionsGteUser(usr, users.USER); err != nil {
		return nil, err
	}

	return o.BaseOperator.SubscribeUserByID(userID, chn)
}

//SubscribeDeviceByID subscribes to all streams of the given device
func (o *AuthOperator) SubscribeDeviceByID(deviceID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	readdev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}

	if _, err := o.permissionsGteDev(readdev, users.DEVICE); err != nil {
		return nil, err
	}

	return o.BaseOperator.SubscribeDeviceByID(deviceID, chn)
}

//SubscribeStreamByID subscribes to the given stream by ID
func (o *AuthOperator) SubscribeStreamByID(streamID int64, substream string, chn chan messenger.Message) (*nats.Subscription, error) {
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

	if dev.RelationToStream(strm, sdevice).Gte(users.DEVICE) {
		return o.BaseOperator.SubscribeStreamByID(streamID, substream, chn)
	}
	return nil, ErrPermissions
}
