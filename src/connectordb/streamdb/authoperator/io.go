package authoperator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/users"
)

//LengthStreamByID returns the length of the stream given its ID
func (o *AuthOperator) LengthStreamByID(streamID int64, substream string) (int64, error) {
	_, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	//If we could read the stream, then we have permissions to get its length
	return o.Db.LengthStreamByID(streamID, substream)
}

//TimeToIndexStreamByID returns the index for the given timestamp
func (o *AuthOperator) TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error) {
	_, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	return o.Db.TimeToIndexStreamByID(streamID, substream, time)
}

//InsertStreamByID inserts into a stream given the stream's ID
func (o *AuthOperator) InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}
	sdevice, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return err
	}

	dev, err := o.Device()
	if err != nil {
		return err
	}

	if !dev.RelationToStream(&strm.Stream, sdevice).Gte(users.DEVICE) {
		return ErrPermissions
	}

	if strm.DeviceId != dev.DeviceId {
		//The writer is not the owner - we set the datastream.Datapoints' sender field
		for i := range data {
			data[i].Sender = o.Name()
		}

		//Since the writer is not the owner, if the stream is a downlink, write to the downlink substream
		if substream == "" && strm.Downlink {
			substream = "downlink"
		}
	} else {
		//The writer is reader. Ensure the sender field is empty
		for i := range data {
			data[i].Sender = ""
		}
	}
	return o.Db.InsertStreamByID(streamID, substream, data, restamp)
}

//GetStreamTimeRangeByID gets the time range by ID
func (o *AuthOperator) GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64) (datastream.DataRange, error) {
	_, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	return o.Db.GetStreamTimeRangeByID(streamID, substream, t1, t2, limit)
}

//GetStreamIndexRangeByID gets an index range by ID
func (o *AuthOperator) GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64) (datastream.DataRange, error) {
	_, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	return o.Db.GetStreamIndexRangeByID(streamID, substream, i1, i2)
}
