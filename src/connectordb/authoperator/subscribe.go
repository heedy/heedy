package authoperator

import (
	"connectordb/messenger"
	"errors"

	"github.com/nats-io/nats"
)

// SubscribeUserByID is not currently supported by AuthOperator
func (a *AuthOperator) SubscribeUserByID(userID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, errors.New("Subscribing by user is currently not supported for authenticated devices")
}

// SubscribeDeviceByID is not currently supported by AuthOperator
func (a *AuthOperator) SubscribeDeviceByID(deviceID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, errors.New("Subscribing by device is currently not supported for authenticated devices")
}

// SubscribeStreamByID subscribes to the given stream
func (a *AuthOperator) SubscribeStreamByID(streamID int64, substream string, chn chan messenger.Message) (*nats.Subscription, error) {
	err := a.ErrorIfNoIOReadAccess(streamID, substream)
	if err != nil {
		return nil, err
	}
	return a.Operator.SubscribeStreamByID(streamID, substream, chn)
}
