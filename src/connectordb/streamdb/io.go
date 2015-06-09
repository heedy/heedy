package streamdb

import (
	"connectordb/streamdb/operator"
	"connectordb/streamdb/timebatchdb"
	"connectordb/streamdb/users"
	"strconv"
)

func getTimebatchUserName(userID int64) string {
	return strconv.FormatInt(userID, 32)
}

func getTimebatchDeviceName(dev *users.Device) string {
	return getTimebatchUserName(dev.UserId) + "/" + strconv.FormatInt(dev.DeviceId, 32)
}

func (o *Database) getStreamTimebatchName(strm *operator.Stream) (string, error) {
	dev, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return "", err
	}
	return getTimebatchDeviceName(dev) + "/" + strconv.FormatInt(strm.StreamId, 32) + "/", nil
}

func (o *Database) getStreamPath(strm *operator.Stream) (string, error) {
	//First try to extract the path from cache
	_, streampath, _ := o.streamCache.GetByID(strm.StreamId)
	if streampath != "" {
		return streampath, nil
	}
	_, devicepath, _ := o.deviceCache.GetByID(strm.DeviceId)
	if devicepath != "" {
		return devicepath + "/" + strm.Name, nil
	}
	//Aight, f this. We need to extract the name - time to pound the database
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
func (o *Database) LengthStreamByID(streamID int64) (int64, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return 0, err
	}
	slen, err := o.tdb.Len(sname)
	return int64(slen), err
}

//TimeToIndexStreamByID returns the index for the given timestamp
func (o *Database) TimeToIndexStreamByID(streamID int64, time float64) (int64, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return 0, err
	}
	sindex, err := o.tdb.GetTimeIndex(sname, IntTimestamp(time))
	return int64(sindex), err
}

//InsertStreamByID inserts into the stream given by the ID
func (o *Database) InsertStreamByID(streamID int64, data []operator.Datapoint, substream string) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}

	dpa, err := strm.ConvertDatapointArray(data)
	if err != nil {
		return err
	}
	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return err
	}
	streampath, err := o.getStreamPath(strm)
	if substream != "" {
		sname = sname + substream
		streampath = streampath + "/" + substream
	}

	//TODO(daniel): We need substream validation code here. This requires the KV store

	if !strm.Ephemeral {
		err = o.tdb.Insert(sname, dpa)
		if err != nil {
			return err
		}
	}

	return o.msg.Publish(sname, operator.Message{streampath, data})
}

//IntTimestamp converts a floating point unix timestamp to nanoseconds
func IntTimestamp(t float64) int64 {
	return int64(1e9 * t)
}

//GetStreamTimeRangeByID reads time range by ID
func (o *Database) GetStreamTimeRangeByID(streamID int64, t1 float64, t2 float64, limit int64, substream string) (operator.DatapointReader, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return nil, err
	}

	//TODO: Substream manipulation code (getting a compute stream by time range and limit is an interpolation
	//query that needs to be messaged to compute handlers, and response waited)

	dr, err := o.tdb.GetTimeRange(sname+substream, IntTimestamp(t1), IntTimestamp(t2))
	if limit > 0 {
		dr = timebatchdb.NewNumRange(dr, uint64(limit))
	}
	return operator.NewRangeReader(dr, strm.GetSchema(), ""), err
}

//GetStreamIndexRangeByID reads index range by ID
func (o *Database) GetStreamIndexRangeByID(streamID int64, i1 int64, i2 int64, substream string) (operator.DatapointReader, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return nil, err
	}

	if i1 < 0 || i2 < 0 || i2 == 0 {
		//We handle negative indices python-style

		streamlength, err := o.LengthStreamByID(streamID)
		if err != nil {
			return nil, err
		}

		if i1 < 0 {
			i1 = streamlength + i1

		}

		if i2 == 0 {
			//For example, getting last datapoint should be (-1,0) in the query
			i2 = streamlength
		} else if i2 < 0 {
			i2 = streamlength + i2
		}

		if i1 < 0 || i2 < 0 {
			//Uh oh - these are still negative. Make the indices 0,0 to fail gracefully
			i1 = 0
			i2 = 0
		}
	}

	dr, err := o.tdb.GetIndexRange(sname+substream, uint64(i1), uint64(i2))
	return operator.NewRangeReader(dr, strm.GetSchema(), ""), err
}
