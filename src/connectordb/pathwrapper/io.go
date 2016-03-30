package pathwrapper

import (
	"connectordb/datastream"
	"util"
)

//LengthStream returns the total number of datapoints in the given stream
func (w Wrapper) LengthStream(streampath string) (int64, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return 0, err
	}
	strm, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return w.LengthStreamByID(strm.StreamID, substream)
}

//TimeToIndexStream returns the index closest to the given timestamp
func (w Wrapper) TimeToIndexStream(streampath string, time float64) (int64, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return 0, err
	}
	strm, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	return w.TimeToIndexStreamByID(strm.StreamID, substream, time)
}

//InsertStream inserts the given array of datapoints into the given stream.
func (w Wrapper) InsertStream(streampath string, data datastream.DatapointArray, restamp bool) error {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return err
	}
	strm, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return err
	}
	return w.InsertStreamByID(strm.StreamID, substream, data, restamp)
}

//GetStreamTimeRange Reads the given stream by time range
func (w Wrapper) GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return w.GetStreamTimeRangeByID(strm.StreamID, substream, t1, t2, limit, transform)
}

//GetShiftedStreamTimeRange Reads the given stream by time range with an index shift
func (w Wrapper) GetShiftedStreamTimeRange(streampath string, t1 float64, t2 float64, shift, limit int64, transform string) (datastream.DataRange, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return w.GetShiftedStreamTimeRangeByID(strm.StreamID, substream, t1, t2, shift, limit, transform)
}

//GetStreamIndexRange Reads the given stream by index range
func (w Wrapper) GetStreamIndexRange(streampath string, i1 int64, i2 int64, transform string) (datastream.DataRange, error) {
	_, _, streampath, _, substream, err := util.SplitStreamPath(streampath)
	if err != nil {
		return nil, err
	}
	strm, err := w.AdminOperator().ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return w.GetStreamIndexRangeByID(strm.StreamID, substream, i1, i2, transform)
}
