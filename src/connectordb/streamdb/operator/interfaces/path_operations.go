package interfaces

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator/messenger"
	"connectordb/streamdb/users"
	"connectordb/streamdb/util"
	"strings"

	"github.com/nats-io/nats"
	"github.com/nu7hatch/gouuid"
)

type PathOperatorMixin struct {
	BaseOperator
}

//SetAdmin does exactly what it claims. It works on both users and devices
func (o *PathOperatorMixin) SetAdmin(path string, isadmin bool) error {
	switch strings.Count(path, "/") {
	default:
		return util.ErrBadPath
	case 0:
		u, err := o.ReadUser(path)
		if err != nil {
			return err
		}
		u.Admin = isadmin
		return o.UpdateUser(u)
	case 1:
		dev, err := o.ReadDevice(path)
		if err != nil {
			return err
		}
		dev.IsAdmin = isadmin
		return o.UpdateDevice(dev)
	}
}

// ReadDevice reads the given device
func (o *PathOperatorMixin) ReadDevice(devicepath string) (*users.Device, error) {
	//Apparently not. Get the device from userdb
	usrname, devname, err := util.SplitDevicePath(devicepath)
	if err != nil {
		return nil, err
	}
	u, err := o.ReadUser(usrname)
	if err != nil {
		return nil, err
	}
	dev, err := o.ReadDeviceByUserID(u.UserId, devname)
	return dev, err
}

//Subscribe given a path, attempts to subscribe to it and its children
func (o *PathOperatorMixin) Subscribe(path string, chn chan messenger.Message) (*nats.Subscription, error) {
	switch strings.Count(path, "/") {
	default:
		return o.SubscribeStream(path, chn)
	case 0:
		return o.SubscribeUser(path, chn)
	case 1:
		return o.SubscribeDevice(path, chn)
	}
}

//ChangeUserPassword changes the password for the given user
func (o *PathOperatorMixin) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	u.SetNewPassword(newpass)
	return o.UpdateUser(u)
}

//DeleteUser deletes a user given the user's name
func (o *PathOperatorMixin) DeleteUser(username string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	return o.DeleteUserByID(u.UserId)
}

//ReadAllDevices for the given user
func (o *PathOperatorMixin) ReadAllDevices(username string) ([]users.Device, error) {
	u, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.ReadAllDevicesByUserID(u.UserId)
}

//CreateDevice creates a new device at the given path
func (o *PathOperatorMixin) CreateDevice(devicepath string) error {
	userName, deviceName, err := util.SplitDevicePath(devicepath)
	if err != nil {
		return err
	}
	u, err := o.ReadUser(userName)
	if err != nil {
		return err
	}

	return o.CreateDeviceByUserID(u.UserId, deviceName)
}

//ChangeDeviceAPIKey generates a new api key for the given device, and returns the key
func (o *PathOperatorMixin) ChangeDeviceAPIKey(devicepath string) (apikey string, err error) {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return "", err
	}
	newkey, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	dev.ApiKey = newkey.String()
	return dev.ApiKey, o.UpdateDevice(dev)
}

//DeleteDevice deletes an existing device
func (o *PathOperatorMixin) DeleteDevice(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err //Workaround for #81
	}
	return o.DeleteDeviceByID(dev.DeviceId)
}

//ReadAllStreams reads all the streams for the given device
func (o *PathOperatorMixin) ReadAllStreams(devicepath string) ([]users.Stream, error) {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	return o.ReadAllStreamsByDeviceID(dev.DeviceId)
}

//CreateStream makes a new stream
func (o *PathOperatorMixin) CreateStream(streampath, jsonschema string) error {
	_, devicepath, _, streamname, _, err := util.SplitStreamPath(streampath)
	if err != nil {
		return err
	}
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err
	}
	return o.CreateStreamByDeviceID(dev.DeviceId, streamname, jsonschema)
}

//DeleteStream deletes the given stream given its path
func (o *PathOperatorMixin) DeleteStream(streampath string) error {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return err
	}
	s, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	return o.DeleteStreamByID(s.StreamId, substream)
}

//LengthStream returns the total number of datapoints in the given stream
func (o *PathOperatorMixin) LengthStream(streampath string) (int64, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return 0, err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return o.LengthStreamByID(strm.StreamId, substream)
}

//TimeToIndexStream returns the index closest to the given timestamp
func (o *PathOperatorMixin) TimeToIndexStream(streampath string, time float64) (int64, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return 0, err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return o.TimeToIndexStreamByID(strm.StreamId, substream, time)
}

//InsertStream inserts the given array of datapoints into the given stream.
func (o *PathOperatorMixin) InsertStream(streampath string, data datastream.DatapointArray, restamp bool) error {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	return o.InsertStreamByID(strm.StreamId, substream, data, restamp)
}

//GetStreamTimeRange Reads the given stream by time range
func (o *PathOperatorMixin) GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64) (datastream.DataRange, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.GetStreamTimeRangeByID(strm.StreamId, substream, t1, t2, limit)
}

//GetStreamIndexRange Reads the given stream by index range
func (o *PathOperatorMixin) GetStreamIndexRange(streampath string, i1 int64, i2 int64) (datastream.DataRange, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.GetStreamIndexRangeByID(strm.StreamId, substream, i1, i2)
}

//SubscribeUser subscribes to everything the user does
func (o *PathOperatorMixin) SubscribeUser(username string, chn chan messenger.Message) (*nats.Subscription, error) {
	usr, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.SubscribeUserByID(usr.UserId, chn)
}

//SubscribeDevice subscribes to everythnig the device does
func (o *PathOperatorMixin) SubscribeDevice(devpath string, chn chan messenger.Message) (*nats.Subscription, error) {
	dev, err := o.ReadDevice(devpath)
	if err != nil {
		return nil, err
	}
	return o.SubscribeDeviceByID(dev.DeviceId, chn)
}

//SubscribeStream subscribes to the given stream
func (o *PathOperatorMixin) SubscribeStream(streampath string, chn chan messenger.Message) (*nats.Subscription, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.SubscribeStreamByID(strm.StreamId, substream, chn)
}

//ReadStream reads the given stream
func (o *PathOperatorMixin) ReadStream(streampath string) (*users.Stream, error) {
	//Make sure that substreams are stripped from read
	_, devicepath, streampath, streamname, _, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}

	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	return o.ReadStreamByDeviceID(dev.DeviceId, streamname)
}
