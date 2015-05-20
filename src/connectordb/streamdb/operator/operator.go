package operator

import (
	"connectordb/streamdb/users"
	"strings"

	"github.com/nu7hatch/gouuid"
)

//Operator defines extension functions which work with any BaseOperator, adding extra functionality.
//In particular, Operator makes querying stuff by name so much easier
type Operator struct {
	BaseOperator
}

//SetAdmin does exactly what it claims. It works on both users and devices
func (o Operator) SetAdmin(path string, isadmin bool) error {
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

//ChangeUserPassword changes the password for the given user
func (o Operator) ChangeUserPassword(username, newpass string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	u.SetNewPassword(newpass)
	return o.UpdateUser(u)
}

//DeleteUser deletes a user given the user's name
func (o Operator) DeleteUser(username string) error {
	u, err := o.ReadUser(username)
	if err != nil {
		return err
	}
	return o.DeleteUserByID(u.UserId)
}

//ReadAllDevices for the given user
func (o Operator) ReadAllDevices(username string) ([]users.Device, error) {
	u, err := o.ReadUser(username)
	if err != nil {
		return nil, err
	}
	return o.ReadAllDevicesByUserID(u.UserId)
}

//CreateDevice creates a new device at the given path
func (o Operator) CreateDevice(devicepath string) error {
	userName, deviceName, err := SplitDevicePath(devicepath, nil)
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
func (o Operator) ChangeDeviceAPIKey(devicepath string) (apikey string, err error) {
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
func (o Operator) DeleteDevice(devicepath string) error {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return err //Workaround for #81
	}
	return o.DeleteDeviceByID(dev.DeviceId)
}

//ReadAllStreams reads all the streams for the given device
func (o Operator) ReadAllStreams(devicepath string) ([]Stream, error) {
	dev, err := o.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	return o.ReadAllStreamsByDeviceID(dev.DeviceId)
}

//CreateStream makes a new stream
func (o Operator) CreateStream(streampath, jsonschema string) error {
	_, devicepath, _, streamname, _, err := SplitStreamPath(streampath, nil)
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
func (o Operator) DeleteStream(streampath string) error {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath, nil)
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
func (o Operator) LengthStream(streampath string) (int64, error) {
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return o.LengthStreamByID(strm.StreamId)
}

//TimeToIndexStream returns the index closest to the given timestamp
func (o Operator) TimeToIndexStream(streampath string, time float64) (int64, error) {
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return o.TimeToIndexStreamByID(strm.StreamId, time)
}

//InsertStream inserts the given array of datapoints into the given stream.
func (o Operator) InsertStream(streampath string, data []Datapoint) error {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath, nil)
	if err != nil {
		return err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	return o.InsertStreamByID(strm.StreamId, data, substream)
}

//GetStreamTimeRange Reads the given stream by time range
func (o Operator) GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64) (DatapointReader, error) {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath, nil)
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.GetStreamTimeRangeByID(strm.StreamId, t1, t2, limit, substream)
}

//GetStreamIndexRange Reads the given stream by index range
func (o Operator) GetStreamIndexRange(streampath string, i1 int64, i2 int64) (DatapointReader, error) {
	_, _, streampath, _, substream, err := SplitStreamPath(streampath, nil)
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.GetStreamIndexRangeByID(strm.StreamId, i1, i2, substream)
}
