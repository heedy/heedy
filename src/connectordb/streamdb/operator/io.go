package operator

import (
	"connectordb/streamdb/datastream"
	"errors"
)

var (
	ErrTimestampOrder = errors.New("Timestamps are not ordered!")
)

func (o *Database) getStreamPath(strm *operator.Stream) (string, error) {
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
func (o *Database) LengthStreamByID(streamID int64, substream string) (int64, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	return o.ds.StreamLength(strm.DeviceId, strm.StreamId, substream)
}

//TimeToIndexStreamByID returns the index for the given timestamp
func (o *Database) TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}

	return o.ds.GetTimeIndex(strm.DeviceId, streamID, substream, time)
}

//InsertStreamByID inserts into the stream given by the ID
func (o *Database) InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error {
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

	return o.msg.Publish(streampath, operator.Message{streampath, data})
}

//GetStreamTimeRangeByID reads time range by ID
func (o *Database) GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64) (datastream.DataRange, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	dr, err := o.ds.TRange(strm.DeviceId, strm.StreamId, substream, t1, t2)
	if limit > 0 {
		dr = datastream.NewNumRange(dr, limit)
	}
	return dr, err
}

//GetStreamIndexRangeByID reads index range by ID
func (o *Database) GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64) (datastream.DataRange, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	return o.ds.IRange(strm.DeviceId, strm.StreamId, substream, i1, i2)
}
