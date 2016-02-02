package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"

	pconfig "config/permissions"
)

// CountStreams returns the total number of users of the entire database
func (a *AuthOperator) CountStreams() (int64, error) {
	perm := pconfig.Get()
	usr, dev, err := a.getUserAndDevice()
	if err != nil {
		return 0, err
	}
	urole := permissions.GetUserRole(perm, usr)
	drole := permissions.GetDeviceRole(perm, dev)
	if !urole.CanCountStreams || !drole.CanCountStreams {
		return 0, errors.New("Don't have permissions necesaary to count streams")
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
		return nil, errors.New("You do not have permissions necessary to list this device's streams.")
	}

	streams, err := a.Operator.ReadAllStreamsByDeviceID(deviceID)
	if err != nil {
		return nil, err
	}

	for i := range streams {
		err = permissions.DeleteDisallowedFields(perm, ua, da, "stream", streams[i])
		if err != nil {
			return nil, err
		}
	}
	return streams, nil
}

// CreateStreamByDeviceID creates the given stream if permitted
func (a *AuthOperator) CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error {
	_, _, _, _, ua, da, err := a.getDeviceAccessLevels(deviceID)
	if err != nil {
		return err
	}

	if !ua.CanCreateStream || !da.CanCreateStream {
		return errors.New("You do not have permissions necessary to create this stream.")
	}

	return a.Operator.CreateStreamByDeviceID(deviceID, streamname, jsonschema)
}

// ReadStreamByID reads the given stream
func (a *AuthOperator) ReadStreamByID(streamID int64) (*users.Stream, error) {
	s, err := a.Operator.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
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

// ReadStreamByDeviceID uses ReadStreamByID internally
func (a *AuthOperator) ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error) {
	stream, err := a.Operator.ReadStreamByDeviceID(deviceID, streamname)
	if err != nil {
		return nil, err
	}
	return a.ReadStreamByID(stream.StreamID)
}

// UpdateStreamByID updates the given stream
func (a *AuthOperator) UpdateStreamByID(streamID int64, updates map[string]interface{}) error {
	s, err := a.Operator.ReadStreamByID(streamID)
	if err != nil {
		return err
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
		return err
	}
	_, _, _, _, ua, da, err := a.getDeviceAccessLevels(s.DeviceID)
	if err != nil {
		return err
	}
	if !ua.CanDeleteStream || !da.CanDeleteStream {
		return errors.New("You do not have permissions necessary to delete this stream.")
	}
	return a.Operator.DeleteStreamByID(streamID, substream)
}
