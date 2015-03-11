package streamdb

import (
    "streamdb/users"
    //"streamdb/dtypes"
    "errors"
    )

type Stream struct {
    users.Stream
    dev *Device
    uri string
}

//Returns the stream object
func (dev *Device) GetStream(streamuri string) (*Stream, error) {
    return nil,nil
}

func (s *Stream) Write([]interface{}) error {
    //Check for write permission of the device to the stream

    //Write using the uri as key to timebatchDB
    return errors.New("UNIMPLEMENTED")
}

func (s *Stream) ReadIndex(i1, i2 uint64) ([]interface{},error) {
    //Check for read permission of the device to the stream

    //Write using the uri as key to timebatchDB
    return nil,errors.New("UNIMPLEMENTED")
}
