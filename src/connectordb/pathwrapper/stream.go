package pathwrapper

import (
	"connectordb/users"
	"util"
)

//ReadDeviceStreams reads all the streams for the given device
func (w Wrapper) ReadDeviceStreams(devicepath string) ([]*users.Stream, error) {
	dev, err := w.AdminOperator().ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	return w.ReadAllStreamsByDeviceID(dev.DeviceID)
}

// ReadUserStreams reads all streams belonging to the user, optionally filtering by public/downlink/visible
func (w Wrapper) ReadUserStreams(username string, public, downlink, hidden bool) ([]*users.DevStream, error) {
	u, err := w.AdminOperator().ReadUser(username)
	if err != nil {
		return nil, err
	}
	return w.ReadAllStreamsByUserID(u.UserID, public, downlink, hidden)
}

//CreateStream makes a new stream
func (w Wrapper) CreateStream(streampath string, s *users.StreamMaker) error {
	_, devicepath, _, streamname, _, err := util.SplitStreamPath(streampath)
	if err != nil {
		return err
	}
	dev, err := w.AdminOperator().ReadDevice(devicepath)
	if err != nil {
		return err
	}
	s.Name = streamname
	s.DeviceID = dev.DeviceID
	return w.CreateStreamByDeviceID(s)
}

//ReadStream reads the given stream
func (w Wrapper) ReadStream(streampath string) (*users.Stream, error) {
	_, devicepath, streampath, streamname, _, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}

	dev, err := w.AdminOperator().ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}
	return w.ReadStreamByDeviceID(dev.DeviceID, streamname)
}

// UpdateStream performs an update on the given stream path
func (w Wrapper) UpdateStream(streampath string, updates map[string]interface{}) error {
	s, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return err
	}

	return w.UpdateStreamByID(s.StreamID, updates)
}

//DeleteStream deletes the given stream given its path
func (w Wrapper) DeleteStream(streampath string) error {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return err
	}
	s, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return err
	}
	return w.DeleteStreamByID(s.StreamID, substream)
}
