package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"
	"fmt"

	pconfig "config/permissions"
)

// CountStreams returns the total number of users of the entire database
func (a *AuthOperator) CountStreams() (int64, error) {
	perm := pconfig.Get()
	usr, dev, err := a.UserAndDevice()
	if err != nil {
		return 0, err
	}
	urole := permissions.GetUserRole(perm, usr)
	drole := permissions.GetDeviceRole(perm, dev)
	if !urole.CanCountStreams || !drole.CanCountStreams {
		return 0, errors.New("Don't have permissions necessary to count streams")
	}
	return a.Operator.CountStreams()
}

// ReadAllStreamsByDeviceID reads all of the streams accessible to the operator
func (a *AuthOperator) ReadAllStreamsByDeviceID(deviceID int64) ([]*users.Stream, error) {
	perm, _, _, _, ua, da, err := a.getDeviceAccessLevels(deviceID)
	if err != nil {
		return nil, err
	}
	if !ua.CanListStreams || !da.CanListStreams {
		return nil, permissions.ErrNoAccess
	}

	streams, err := a.Operator.ReadAllStreamsByDeviceID(deviceID)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}

	for i := range streams {
		err = permissions.DeleteDisallowedFields(perm, ua, da, "stream", streams[i])
		if err != nil {
			return nil, err
		}
	}
	return streams, nil
}

// ReadDeviceStreamsToMap reads all of the streams who this device has permissions to read to a map
func (a *AuthOperator) ReadDeviceStreamsToMap(devname string) ([]map[string]interface{}, error) {
	dev, err := a.Operator.ReadDevice(devname)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	_, _, _, _, ua, da, err := a.getDeviceAccessLevels(dev.DeviceID)
	if err != nil {
		return nil, err
	}
	if !ua.CanListStreams || !da.CanListStreams {
		return nil, permissions.ErrNoAccess
	}

	// See ReadAllUsers
	ss, err := a.Operator.ReadDeviceStreams(devname)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	result := make([]map[string]interface{}, 0, len(ss))
	for i := range ss {
		u, err := a.ReadStreamToMap(devname + "/" + ss[i].Name)
		if err == nil {
			result = append(result, u)
		}
	}
	return result, nil
}

// StreamMaker returns the StreamMaker prepopulated with default values
// TODO: This is a hack - it does not set defaults for subdevices
// and substreams. Furthermore, create allows setting ALL properties,
// which is definitely not wanted
func (a *AuthOperator) StreamMaker() (*users.StreamMaker, error) {
	u, err := a.User()
	if err != nil {
		return nil, err
	}
	perm := pconfig.Get()
	// Make sure that the given role exists
	r, ok := perm.UserRoles[u.Role]
	if !ok {
		return nil, fmt.Errorf("The given role '%s' does not exist", u.Role)
	}

	d := r.CreateStreamDefaults
	return &users.StreamMaker{
		Stream: users.Stream{
			Nickname:    d.Nickname,
			Description: d.Description,
			Icon:        d.Icon,
			Schema:      d.Schema,
			Datatype:    d.Datatype,
			Ephemeral:   d.Ephemeral,
			Downlink:    d.Downlink,
		},
	}, nil
}

// CreateStreamByDeviceID creates the given stream if permitted
func (a *AuthOperator) CreateStreamByDeviceID(sm *users.StreamMaker) error {
	_, _, _, _, ua, da, err := a.getDeviceAccessLevels(sm.DeviceID)
	if err != nil {
		return err
	}

	if !ua.CanCreateStream || !da.CanCreateStream {
		return permissions.ErrNoAccess
	}

	return a.Operator.CreateStreamByDeviceID(sm)
}

// ReadStreamByID reads the given stream
func (a *AuthOperator) ReadStreamByID(streamID int64) (*users.Stream, error) {
	s, err := a.Operator.ReadStreamByID(streamID)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	perm, _, _, _, ua, da, err := a.getDeviceAccessLevels(s.DeviceID)
	if err != nil {
		return nil, err
	}
	err = permissions.DeleteDisallowedFields(perm, ua, da, "stream", s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// ReadStreamToMap reads the given stream into a map, where only the permitted fields are present in the map
func (a *AuthOperator) ReadStreamToMap(spath string) (map[string]interface{}, error) {
	s, err := a.Operator.ReadStream(spath)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	perm, _, _, _, ua, da, err := a.getDeviceAccessLevels(s.DeviceID)
	if err != nil {
		return nil, err
	}
	return permissions.ReadObjectToMap(perm, ua, da, "stream", s)
}

// ReadStreamByDeviceID uses ReadStreamByID internally
func (a *AuthOperator) ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error) {
	stream, err := a.Operator.ReadStreamByDeviceID(deviceID, streamname)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	return a.ReadStreamByID(stream.StreamID)
}

// UpdateStreamByID updates the given stream
func (a *AuthOperator) UpdateStreamByID(streamID int64, updates map[string]interface{}) error {
	s, err := a.Operator.ReadStreamByID(streamID)
	if err != nil {
		return permissions.ErrNoAccess
	}
	perm, _, _, _, ua, da, err := a.getDeviceAccessLevels(s.DeviceID)
	if err != nil {
		return err
	}
	err = permissions.CheckIfUpdateFieldsPermitted(perm, ua, da, "stream", updates)
	if err != nil {
		return err
	}
	return a.Operator.UpdateStreamByID(streamID, updates)
}

// DeleteStreamByID deletes the given stream
func (a *AuthOperator) DeleteStreamByID(streamID int64, substream string) error {
	s, err := a.Operator.ReadStreamByID(streamID)
	if err != nil {
		return permissions.ErrNoAccess
	}
	_, _, _, _, ua, da, err := a.getDeviceAccessLevels(s.DeviceID)
	if err != nil {
		return err
	}
	if !ua.CanDeleteStream || !da.CanDeleteStream {
		return permissions.ErrNoAccess
	}
	return a.Operator.DeleteStreamByID(streamID, substream)
}
