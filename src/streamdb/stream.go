package streamdb

import (
    "streamdb/users"
    "streamdb/dtypes"
    "strings"
    "errors"
    )

type Stream struct {
    Stream *users.Stream
    dev *Device
    uri string
}



//Returns the stream object
func (dev *Device) GetStream(streamuri string) (*Stream, error) {
    uds := strings.Split(streamuri,"/")
    if len(uds)!=3 {
        return nil,errors.New("Could not get stream: incorrect number of arguments.")
    }
    _,_,s,err :=dev.db.ReadStreamByUriAs(dev.Device,uds[0],uds[1],uds[2])

    if err!=nil {
        return nil,err
    }

    return &Stream{s,dev,streamuri},nil
}

func (s *Stream) Write(pt dtypes.TypedDatapoint) error {
    //Check for write permission of the device to the stream
    stream := s.Stream

    if HasPermissions(s.dev.Device, write_privilege) && s.Stream.OwnerId == s.dev.Device.Id{
        return s.dev.db.tdb.Insert(pt)
    }

    if HasPermissions(s.dev.Device, super_privilege) {
        return s.dev.db.tdb.Insert(pt)
    }

    owner, err := s.dev.db.ReadStreamOwner(stream.Id) // user
    if err != nil {
        return err
    }

    if s.dev.Device.OwnerId == owner.Id && HasPermissions(s.dev.Device, write_anywhere_privilege) {
        return s.dev.db.tdb.Insert(pt)
    }

    return nil
}

func canReadStream(dev *users.Device, stream *users.Stream, db *Database) (bool, error) {
    if HasPermissions(dev, super_privilege) {
        return true, nil
    }

    if HasPermissions(dev, read_privilege) && stream.OwnerId == dev.Id {
        return true, nil
    }

    owner, err := db.ReadStreamOwner(stream.Id) // user

    if err != nil {
        return false, err
    }

    if dev.OwnerId == owner.Id{
        return true, nil
    }

    return false, nil
}

func (s *Stream) ReadIndex(i1, i2 uint64) (d *dtypes.TypedRange,err error) {
    //Check for read permission of the device to the stream

    read, err := canReadStream(s.dev.Device, s.Stream, s.dev.db)

    if err != nil {
        return nil, err
    }

    if ! read {
        return nil, PrivilegeError
    }

    //Write using the uri as key to timebatchDB
    tr, err := s.dev.db.tdb.GetIndexRange(s.uri,s.Stream.Type,i1,i2),nil
    return &tr, err
}


func (s *Stream) ReadTime(t1, t2 int64) (d *dtypes.TypedRange,err error) {
    //Check for read permission of the device to the stream

    read, err := canReadStream(s.dev.Device, s.Stream, s.dev.db)

    if err != nil {
        return nil, err
    }

    if ! read {
        return nil, PrivilegeError
    }
    //Write using the uri as key to timebatchDB
    tr, err := s.dev.db.tdb.GetTimeRange(s.uri,s.Stream.Type,t1,t2),nil
    return &tr, err
}

func (s *Stream) EmptyDatapoint() dtypes.TypedDatapoint {
    d,ok :=dtypes.GetType(s.Stream.Type)
    if !ok {
        return nil
    }
    return d.New()
}
