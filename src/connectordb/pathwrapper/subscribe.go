package pathwrapper

import (
	"connectordb/messenger"
	"strings"
	"util"

	"github.com/nats-io/nats"
)

//SubscribeUser subscribes to everything the user does
func (w Wrapper) SubscribeUser(username string, chn chan messenger.Message) (*nats.Subscription, error) {
	usr, err := w.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return w.SubscribeUserByID(usr.UserID, chn)
}

//SubscribeDevice subscribes to everythnig the device does
func (w Wrapper) SubscribeDevice(devpath string, chn chan messenger.Message) (*nats.Subscription, error) {
	dev, err := w.ReadDevice(devpath)
	if err != nil {
		return nil, err
	}
	return w.SubscribeDeviceByID(dev.DeviceID, chn)
}

//SubscribeStream subscribes to the given stream
func (w Wrapper) SubscribeStream(streampath string, chn chan messenger.Message) (*nats.Subscription, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := w.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return w.SubscribeStreamByID(strm.StreamID, substream, chn)
}

//Subscribe given a path, attempts to subscribe to it and its children
func (w Wrapper) Subscribe(path string, chn chan messenger.Message) (*nats.Subscription, error) {
	switch strings.Count(path, "/") {
	default:
		return w.SubscribeStream(path, chn)
	case 0:
		return w.SubscribeUser(path, chn)
	case 1:
		return w.SubscribeDevice(path, chn)
	}
}
