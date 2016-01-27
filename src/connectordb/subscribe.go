/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package connectordb

import (
	"connectordb/messenger"

	"github.com/nats-io/nats"
)

//SubscribeUserByID subscribes to everything a user creates
func (db *Database) SubscribeUserByID(userID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	usr, err := db.ReadUserByID(userID)
	if err != nil {
		return nil, err
	}
	return db.msg.Subscribe(usr.Name+"/*/*", chn)
}

//SubscribeDeviceByID subscribes to all streams of the given device
func (db *Database) SubscribeDeviceByID(deviceID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	dev, err := db.ReadDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	usr, err := db.ReadUserByID(dev.UserID)
	if err != nil {
		return nil, err
	}
	return db.msg.Subscribe(usr.Name+"/"+dev.Name+"/*", chn)
}

//SubscribeStreamByID subscribes to the given stream by ID
func (db *Database) SubscribeStreamByID(streamID int64, substream string, chn chan messenger.Message) (*nats.Subscription, error) {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	_, _, routing, err := db.getStreamPath(strm)
	if err != nil {
		return nil, err
	}
	if substream != "" {
		routing = routing + "/" + substream
	}
	return db.msg.Subscribe(routing, chn)
}
