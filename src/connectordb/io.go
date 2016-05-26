/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package connectordb

import (
	pconfig "config/permissions"
	"connectordb/authoperator/permissions"
	"connectordb/datastream"
	"connectordb/messenger"
	"connectordb/query"
	"connectordb/users"
	"errors"
)

var (
	// ErrTimestampOrder is thrown when the tiemstamps are not increasing
	ErrTimestampOrder = errors.New("Timestamps are not ordered!")
)

func (db *Database) getStreamPath(strm *users.Stream) (*users.User, *users.Device, string, error) {
	dev, err := db.ReadDeviceByID(strm.DeviceID)
	if err != nil {
		return nil, nil, "", err
	}
	usr, err := db.ReadUserByID(dev.UserID)
	if err != nil {
		return nil, nil, "", err
	}
	return usr, dev, usr.Name + "/" + dev.Name + "/" + strm.Name, nil
}

//LengthStreamByID returns the total number of datapoints in the stream by ID
func (db *Database) LengthStreamByID(streamID int64, substream string) (int64, error) {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	return db.DataStream.StreamLength(strm.DeviceID, strm.StreamID, substream)
}

//TimeToIndexStreamByID returns the index for the given timestamp
func (db *Database) TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error) {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}

	return db.DataStream.GetTimeIndex(strm.DeviceID, streamID, substream, time)
}

//InsertStreamByID inserts into the stream given by the ID
func (db *Database) InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error {
	strm, err := db.ReadStreamByID(streamID)
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

	u, _, streampath, err := db.getStreamPath(strm)
	if substream != "" {
		streampath = streampath + "/" + substream
	}
	if err != nil {
		return err
	}

	if !strm.Ephemeral {

		r := permissions.GetUserRole(pconfig.Get(), u)
		_, err = db.DataStream.Insert(strm.DeviceID, strm.StreamID, substream, data, restamp, r.MaxDeviceSize, r.MaxStreamSize)
		if err != nil {
			return err
		}
	}

	return db.Messenger.Publish(streampath, messenger.Message{streampath, "", data})
}

//GetStreamTimeRangeByID reads time range by ID
func (db *Database) GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error) {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	dr, err := db.DataStream.TRange(strm.DeviceID, strm.StreamID, substream, t1, t2)
	if limit > 0 {
		dr = datastream.NewNumRange(dr, limit)
	}
	//Add a transform to the resulting data range if one is wanted
	if transform != "" {
		tr, err := query.NewExtendedTransformRange(dr, transform)
		if err != nil {
			dr.Close()
			return nil, err
		}
		dr = tr
	}

	return dr, err
}

//GetShiftedStreamTimeRangeByID reads time range by ID with an index shift
func (db *Database) GetShiftedStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, shift, limit int64, transform string) (datastream.DataRange, error) {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	dr, err := db.DataStream.TimePlusIndexRange(strm.DeviceID, strm.StreamID, substream, t1, t2, shift)
	if err != nil {
		return nil, err
	}

	if limit > 0 {
		dr = datastream.NewNumRange(dr, limit)
	}
	//Add a transform to the resulting data range if one is wanted
	if transform != "" {
		tr, err := query.NewExtendedTransformRange(dr, transform)
		if err != nil {
			dr.Close()
			return nil, err
		}
		dr = tr
	}

	return dr, err
}

//GetStreamIndexRangeByID reads index range by ID
func (db *Database) GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64, transform string) (datastream.DataRange, error) {
	strm, err := db.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	dr, err := db.DataStream.IRange(strm.DeviceID, strm.StreamID, substream, i1, i2)
	if err != nil {
		return nil, err
	}

	//Add a transform to the resulting data range if one is wanted
	if transform != "" {
		tr, err := query.NewExtendedTransformRange(dr, transform)
		if err != nil {
			dr.Close()
			return nil, err
		}
		dr = tr
	}
	return dr, err
}
