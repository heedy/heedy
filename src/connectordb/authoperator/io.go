package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/datastream"
	"errors"

	pconfig "config/permissions"
)

func (a *AuthOperator) getIOPermissions(streamID int64) (*pconfig.Permissions, *pconfig.AccessLevel, *pconfig.AccessLevel, error) {
	s, err := a.Operator.ReadStreamByID(streamID)
	if err != nil {
		return nil, nil, nil, err
	}
	perm, _, _, _, ua, da, err := a.getDeviceAccessLevels(s.DeviceID)
	if err != nil {
		return nil, nil, nil, err
	}
	return perm, ua, da, nil
}

// ErrorIfNoIOReadAccess returns the permissions for reading the given stream
func (a *AuthOperator) ErrorIfNoIOReadAccess(streamID int64) error {
	perm, ua, da, err := a.getIOPermissions(streamID)
	if err != nil {
		return err
	}
	if !permissions.GetReadAccess(perm, ua).CanAccessStreamData || !permissions.GetReadAccess(perm, da).CanAccessStreamData {
		return errors.New("Access to stream data denied.")
	}
	return nil
}

// LengthStreamByID gets the stream's length
func (a *AuthOperator) LengthStreamByID(streamID int64, substream string) (int64, error) {
	err := a.ErrorIfNoIOReadAccess(streamID)
	if err != nil {
		return 0, err
	}
	return a.Operator.LengthStreamByID(streamID, substream)
}

// TimeToIndexStreamByID gets the time to index. more documentatino in definition of Operator
func (a *AuthOperator) TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error) {
	err := a.ErrorIfNoIOReadAccess(streamID)
	if err != nil {
		return 0, err
	}
	return a.Operator.TimeToIndexStreamByID(streamID, substream, time)
}

// InsertStreamByID inserts the given data into the stream
func (a *AuthOperator) InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error {
	perm, ua, da, err := a.getIOPermissions(streamID)
	if err != nil {
		return err
	}
	if !permissions.GetWriteAccess(perm, ua).CanAccessStreamData || !permissions.GetWriteAccess(perm, da).CanAccessStreamData {
		return errors.New("Write access to stream data denied.")
	}
	strm, err := a.Operator.ReadStreamByID(streamID)
	if err != nil {
		return err
	}
	dev, err := a.Device()
	if err != nil {
		return err
	}
	if dev.DeviceID != strm.DeviceID {
		//The writer is not the owner - we set the datastream.Datapoints' sender field
		for i := range data {
			data[i].Sender = a.Name()
		}

		//Since the writer is not the owner, if the stream is a downlink, write to the downlink substream
		if strm.Downlink && substream == "" {
			substream = "downlink"
		}
	} else {
		//The writer is reader. Ensure the sender field is empty
		for i := range data {
			data[i].Sender = ""
		}
	}
	return a.Operator.InsertStreamByID(streamID, substream, data, restamp)
}

// GetStreamTimeRangeByID is defined in Operator
func (a *AuthOperator) GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error) {
	err := a.ErrorIfNoIOReadAccess(streamID)
	if err != nil {
		return nil, err
	}
	return a.Operator.GetStreamTimeRangeByID(streamID, substream, t1, t2, limit, transform)
}

// GetStreamIndexRangeByID is defined in Operator
func (a *AuthOperator) GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64, transform string) (datastream.DataRange, error) {
	err := a.ErrorIfNoIOReadAccess(streamID)
	if err != nil {
		return nil, err
	}
	return a.Operator.GetStreamIndexRangeByID(streamID, substream, i1, i2, transform)
}

// GetShiftedStreamTimeRangeByID is defined in Operator
func (a *AuthOperator) GetShiftedStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, shift, limit int64, transform string) (datastream.DataRange, error) {
	err := a.ErrorIfNoIOReadAccess(streamID)
	if err != nil {
		return nil, err
	}
	return a.Operator.GetShiftedStreamTimeRangeByID(streamID, substream, t1, t2, shift, limit, transform)
}
