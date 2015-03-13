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

    return s.dev.db.tdb.Insert(pt)

}

func (s *Stream) ReadIndex(i1, i2 uint64) (d dtypes.TypedRange,err error) {
    //Check for read permission of the device to the stream

    //Write using the uri as key to timebatchDB
    return s.dev.db.tdb.GetIndexRange(s.uri,s.Stream.Type,i1,i2),nil
}


func (s *Stream) ReadTime(t1, t2 int64) (d dtypes.TypedRange,err error) {
    //Check for read permission of the device to the stream

    //Write using the uri as key to timebatchDB
    return s.dev.db.tdb.GetTimeRange(s.uri,s.Stream.Type,t1,t2),nil
}

func (s *Stream) EmptyDatapoint() dtypes.TypedDatapoint {
    d,ok :=dtypes.GetType(s.Stream.Type)
    if !ok {
        return nil
    }
    return d.New()
}
