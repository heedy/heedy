package operator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/users"
	"strings"

	"github.com/nats-io/nats"
	"github.com/nu7hatch/gouuid"
)

//Operator defines extension functions which work with any BaseOperatorInterface, adding extra functionality.
//In particular, Operator makes querying stuff by name so much easier
type Operator struct {
	BaseOperatorInterface
}

//SetAdmin does exactly what it claims. It works on both users and devices
func (o *Operator) SetAdmin(path string, isadmin bool) error {
	switch strings.Count(path, "/") {
	default:
		return ErrBadPath
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

//Subscribe given a path, attempts to subscribe to it and its childrens
func (o *Operator) Subscribe(path string, chn chan Message) (*nats.Subscription, error) {
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
func (o *Operator) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	u.SetNewPassword(newpass)
	return o.UpdateUser(u)
}

//DeleteUser deletes a user given the user's name
func (o *Operator) DeleteUser(username string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	return o.DeleteUserByID(u.UserId)
}

//ReadAllDevices for the given user
func (o *Operator) ReadAllDevices(username string) ([]users.Device, error) {
	u, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.ReadAllDevicesByUserID(u.UserId)
}

//CreateDevice creates a new device at the given path
func (o *Operator) CreateDevice(devicepath string) error {
	userName, deviceName, err := SplitDevicePath(devicepath)
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
func (o *Operator) ChangeDeviceAPIKey(devicepath string) (apikey string, err error) {
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
func (o *Operator) DeleteDevice(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err //Workaround for #81
	}
	return o.DeleteDeviceByID(dev.DeviceId)
}

//ReadAllStreams reads all the streams for the given device
func (o *Operator) ReadAllStreams(devicepath string) ([]Stream, error) {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	return o.ReadAllStreamsByDeviceID(dev.DeviceId)
}

//CreateStream makes a new stream
func (o *Operator) CreateStream(streampath, jsonschema string) error {
	_, devicepath, _, streamname, _, err := SplitStreamPath(streampath)
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
func (o *Operator) DeleteStream(streampath string) error {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath)
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
func (o *Operator) LengthStream(streampath string) (int64, error) {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath)
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
func (o *Operator) TimeToIndexStream(streampath string, time float64) (int64, error) {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath)
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
func (o *Operator) InsertStream(streampath string, data datastream.DatapointArray, restamp bool) error {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath)
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
func (o *Operator) GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64) (datastream.DataRange, error) {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath)
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
func (o *Operator) GetStreamIndexRange(streampath string, i1 int64, i2 int64) (datastream.DataRange, error) {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath)
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
func (o *Operator) SubscribeUser(username string, chn chan Message) (*nats.Subscription, error) {
	usr, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.SubscribeUserByID(usr.UserId, chn)
}

//SubscribeDevice subscribes to everythnig the device does
func (o *Operator) SubscribeDevice(devpath string, chn chan Message) (*nats.Subscription, error) {
	dev, err := o.ReadDevice(devpath)
	if err != nil {
		return nil, err
	}
	return o.SubscribeDeviceByID(dev.DeviceId, chn)
}

//SubscribeStream subscribes to the given stream
func (o *Operator) SubscribeStream(streampath string, chn chan Message) (*nats.Subscription, error) {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.SubscribeStreamByID(strm.StreamId, substream, chn)
}
