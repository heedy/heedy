package streamdb

import (
    "streamdb/users"
    //"streamdb/dtypes"
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
    return nil
}
