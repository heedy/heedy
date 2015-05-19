package streamdb

import (
	"connectordb/streamdb/timebatchdb"
	"connectordb/streamdb/users"
	"connectordb/streamdb/util"
	"strconv"
)

func getTimebatchUserName(userID int64) string {
	return strconv.FormatInt(userID, 32)
}

func getTimebatchDeviceName(dev *users.Device) string {
	return getTimebatchUserName(dev.UserId) + "/" + strconv.FormatInt(dev.DeviceId, 32)
}

func (o *Database) getStreamTimebatchName(strm *Stream) (string, error) {
	dev, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return "", err
	}
	return getTimebatchDeviceName(dev) + "/" + strconv.FormatInt(strm.StreamId, 32) + "/", nil
}

//LengthStream returns the total number of datapoints in the given stream
func (o *Database) LengthStream(streampath string) (int64, error) {
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return o.LengthStreamByID(strm.StreamId)
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

//TimeToIndexStream returns the index closest to the given timestamp
func (o *Database) TimeToIndexStream(streampath string, time float64) (int64, error) {
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return o.TimeToIndexStreamByID(strm.StreamId, time)
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

//InsertStream inserts the given array of datapoints into the given stream.
func (o *Database) InsertStream(streampath string, data []Datapoint) error {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath, nil)
	if err != nil {
		return err
	}
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	return o.InsertStreamByID(strm.StreamId, data, substream)
}

//InsertStreamByID inserts into the stream given by the ID
func (o *Database) InsertStreamByID(streamID int64, data []Datapoint, substream string) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}

	dpa, err := strm.convertDatapointArray(data)
	if err != nil {
		return err
	}
	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return err
	}

	//TODO(daniel): We need substream validation code here. This requires the KV store

	return o.tdb.Insert(sname+substream, dpa)
}

//IntTimestamp converts a floating point unix timestamp to nanoseconds
func IntTimestamp(t float64) int64 {
	return int64(1e9 * t)
}

//GetStreamTimeRange Reads the given stream by time range
func (o *Database) GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64) (DatapointReader, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath, nil)
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.GetStreamTimeRangeByID(strm.StreamId, t1, t2, limit, substream)
}

//GetStreamTimeRangeByID reads time range by ID
func (o *Database) GetStreamTimeRangeByID(streamID int64, t1 float64, t2 float64, limit int64, substream string) (DatapointReader, error) {
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
	return NewRangeReader(dr, strm.s, ""), err
}

//GetStreamIndexRange Reads the given stream by index range
func (o *Database) GetStreamIndexRange(streampath string, i1 int64, i2 int64) (DatapointReader, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath, nil)
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.GetStreamIndexRangeByID(strm.StreamId, i1, i2, substream)
}

//GetStreamIndexRangeByID reads index range by ID
func (o *Database) GetStreamIndexRangeByID(streamID int64, i1 int64, i2 int64, substream string) (DatapointReader, error) {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}

	sname, err := o.getStreamTimebatchName(strm)
	if err != nil {
		return nil, err
	}

	dr, err := o.tdb.GetIndexRange(sname+substream, uint64(i1), uint64(i2))
	return NewRangeReader(dr, strm.s, ""), err
}
