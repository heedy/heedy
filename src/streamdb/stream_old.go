package streamdb

/*
import (
	"errors"
	"streamdb/dtypes"
	"streamdb/users"
	"strings"
)

type StreamOperator struct {
	Stream *users.Stream
	Dev    *Operator
	Uri    string
}

//Returns the stream object
func (dev *Operator) GetStream(streamuri string) (*StreamOperator, error) {
	uds := strings.Split(streamuri, "/")
	if len(uds) != 3 {
		return nil, errors.New("Could not get stream: incorrect number of arguments.")
	}
	_, _, s, err := dev.ReadStreamByUri(uds[0], uds[1], uds[2])

	if err != nil {
		return nil, err
	}

	return &StreamOperator{s, dev, streamuri}, nil
}


func (s *StreamOperator) Write(pt dtypes.TypedDatapoint) error {
	//Check for write permission of the device to the stream
	stream := s.Stream

	if HasPermissions(s.Dev.GetDevice(), write_privilege) && s.Stream.DeviceId == s.Dev.GetDevice().DeviceId {
		return s.Dev.GetDatabase().tdb.InsertKey(s.Uri, pt)
	}

	if HasPermissions(s.Dev.GetDevice(), super_privilege) {
		return s.Dev.GetDatabase().tdb.InsertKey(s.Uri, pt)
	}

	owner, err := s.Dev.GetDatabase().ReadStreamOwner(stream.StreamId) // user
	if err != nil {
		return err
	}

	if s.Dev.GetDevice().UserId == owner.UserId && HasPermissions(s.Dev.GetDevice(), write_anywhere_privilege) {
		return s.Dev.GetDatabase().tdb.InsertKey(s.Uri, pt)
	}

	return nil
}

func canReadStream(dev *users.Device, stream *users.Stream, db *Database) (bool, error) {
	if HasPermissions(dev, super_privilege) {
		return true, nil
	}

	if HasPermissions(dev, read_privilege) && stream.DeviceId == dev.DeviceId {
		return true, nil
	}

	owner, err := db.ReadStreamOwner(stream.StreamId) // user

	if err != nil {
		return false, err
	}

	if dev.UserId == owner.UserId {
		return true, nil
	}

	return false, nil
}

func (s *StreamOperator) ReadIndex(i1, i2 uint64) (d *dtypes.TypedRange, err error) {
	//Check for read permission of the device to the stream

	read, err := canReadStream(s.Dev.GetDevice(), s.Stream, s.Dev.GetDatabase())

	if err != nil {
		return nil, err
	}

	if !read {
		return nil, PrivilegeError
	}

	//Write using the uri as key to timebatchDB
	tr, err := s.Dev.GetDatabase().tdb.GetIndexRange(s.Uri, s.Stream.Type, i1, i2), nil
	return &tr, err
}

func (s *StreamOperator) ReadTime(t1, t2 int64) (d *dtypes.TypedRange, err error) {
	//Check for read permission of the device to the stream

	read, err := canReadStream(s.Dev.GetDevice(), s.Stream, s.Dev.GetDatabase())

	if err != nil {
		return nil, err
	}

	if !read {
		return nil, PrivilegeError
	}
	//Write using the uri as key to timebatchDB
	tr, err := s.Dev.GetDatabase().tdb.GetTimeRange(s.Uri, s.Stream.Type, t1, t2), nil
	return &tr, err
}

func (s *StreamOperator) EmptyDatapoint() dtypes.TypedDatapoint {
	d, ok := dtypes.GetType(s.Stream.Type)
	if !ok {
		return nil
	}
	return d.New()
}
*/
