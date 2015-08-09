package plainoperator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator/messenger"
	"connectordb/streamdb/query"
	"connectordb/streamdb/users"
	"errors"
)

var (
	ErrTimestampOrder = errors.New("Timestamps are not ordered!")
)

func (o *PlainOperator) getStreamPath(strm *users.Stream) (string, error) {
	dev, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return "", err
	}
	usr, err := o.ReadUserByID(dev.UserId)
	if err != nil {
		return "", err
	}
	return usr.Name + "/" + dev.Name + "/" + strm.Name, nil
}

//LengthStreamByID returns the total number of datapoints in the stream by ID
func (o *PlainOperator) LengthStreamByID(streamID int64, substream string) (int64, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	return o.ds.StreamLength(strm.DeviceId, strm.StreamId, substream)
}

//TimeToIndexStreamByID returns the index for the given timestamp
func (o *PlainOperator) TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}

	return o.ds.GetTimeIndex(strm.DeviceId, streamID, substream, time)
}

//InsertStreamByID inserts into the stream given by the ID
func (o *PlainOperator) InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}
	data.SetZeroTime()
	//Now check that everything is okay
	if !strm.Validate(data) {
		return datastream.ErrInvalidDatapoint
	}
	if !data.IsTimestampOrdered() {
		return ErrTimestampOrder
	}

	streampath, err := o.getStreamPath(strm)
	if substream != "" {
		streampath = streampath + "/" + substream
	}
	if err != nil {
		return err
	}

	if !strm.Ephemeral {
		_, err = o.ds.Insert(strm.DeviceId, strm.StreamId, substream, data, restamp)
		if err != nil {
			return err
		}
	}

	return o.msg.Publish(streampath, messenger.Message{streampath, data})
}

//GetStreamTimeRangeByID reads time range by ID
func (o *PlainOperator) GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	dr, err := o.ds.TRange(strm.DeviceId, strm.StreamId, substream, t1, t2)
	if limit > 0 {
		dr = datastream.NewNumRange(dr, limit)
	}
	//Add a transform to the resulting data range if one is wanted
	if transform != "" {
		tr, err := query.NewStreamTransformRange(dr, transform)
		if err != nil {
			dr.Close()
			return nil, err
		}
		dr = tr
	}

	return dr, err
}

//GetShiftedStreamTimeRangeByID reads time range by ID with an index shift
func (o *PlainOperator) GetShiftedStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, shift, limit int64, transform string) (datastream.DataRange, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	dr, err := o.ds.TimePlusIndexRange(strm.DeviceId, strm.StreamId, substream, t1, t2, shift)
	if err != nil {
		return nil, err
	}

	if limit > 0 {
		dr = datastream.NewNumRange(dr, limit)
	}
	//Add a transform to the resulting data range if one is wanted
	if transform != "" {
		tr, err := query.NewStreamTransformRange(dr, transform)
		if err != nil {
			dr.Close()
			return nil, err
		}
		dr = tr
	}

	return dr, err
}

//GetStreamIndexRangeByID reads index range by ID
func (o *PlainOperator) GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64, transform string) (datastream.DataRange, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	dr, err := o.ds.IRange(strm.DeviceId, strm.StreamId, substream, i1, i2)
	if err != nil {
		return nil, err
	}

	//Add a transform to the resulting data range if one is wanted
	if transform != "" {
		tr, err := query.NewStreamTransformRange(dr, transform)
		if err != nil {
			dr.Close()
			return nil, err
		}
		dr = tr
	}
	return dr, err
}
