package users

import (
    "testing"
    "reflect"
    "log"
)

var (
    carid int64
    usrid int64
    usr *User
    devid int64
    dev *Device
)


func init() {
    var err error
    carid, err = CreatePhoneCarrier("StreamTestPhoneCarrier", "StreamTestPhoneCarrier.com")
    if err != nil {
        log.Print(err)
    }

    usrid, err = CreateUser("StreamTestUserName", "StreamTestUserEmail", "StreamTestUserPassword")
    if err != nil {
        log.Print(err)
    }

    usr, err = ReadUserById(usrid)
    if err != nil {
        log.Print(err)
    }

    devid, err = CreateDevice("StreamTestDevice", usr)
    if err != nil {
        log.Print(err)
    }

    dev, err = ReadDeviceById(devid)
    if err != nil {
        log.Print(err)
    }

}


func TestCreateStream(t *testing.T) {
    _, err := CreateStream("TestCreateStream", "{}", "{}", dev)
    if(err != nil) {
        t.Errorf("Cannot create stream %v", err)
        return
    }

    _, err = CreateStream("TestCreateStream", "{}", "{}", dev)
    if(err == nil) {
        t.Errorf("Created stream with duplicate name")
    }
}

func TestReadStreamById(t *testing.T) {
    streamid, err := CreateStream("TestReadStreamById", "{}", "{}", dev)
    if(err != nil) {
        t.Errorf("Cannot create stream %v", err)
        return
    }

    stream, err := ReadStreamById(streamid)

    if err != nil || stream == nil {
        t.Errorf("Cannot read stream back with returned id %v", streamid)
        return
    }
}

func TestUpdateStream(t *testing.T) {
    streamid, err := CreateStream("TestUpdateStream", "{}", "{}", dev)
    if(err != nil) {
        t.Errorf("Cannot create stream %v", err)
        return
    }

    stream, err := ReadStreamById(streamid)

    if err != nil || stream == nil{
        t.Errorf("Cannot read stream back with returned id %v", streamid)
        return
    }

    stream.Name = "A"
    stream.Active = false
    stream.Public = true
    stream.Schema_Json = "{a:'string'}"
    stream.Defaults_Json = "{a:'b'}"
    //stream.OwnerId = dev

    err = UpdateStream(stream)

    if err != nil {
        t.Errorf("Could not update stream %v", err)
    }

    stream2, err := ReadStreamById(streamid)

    if err != nil {
        t.Errorf("got an error when trying to get a stream that should exist %v", err)
        return
    }

    if ! reflect.DeepEqual(stream, stream2) {
        t.Errorf("The original and updated objects don't match orig: %v updated %v", stream, stream2)
    }
}



func TestDeleteStream(t *testing.T) {
    id, err := CreateStream("TestDeleteStream", "{}", "{}", dev)

    if nil != err {
        t.Errorf("Cannot create stream to test delete")
        return
    }

    err = DeleteStream(id)

    if nil != err {
        t.Errorf("Error when attempted delete %v", err)
        return
    }

    stream, err := ReadStreamById(id)

    if err == nil {
        t.Errorf("The stream with the selected ID should have errored out, but it was not")
        return
    }

    if stream != nil {
        t.Errorf("Expected nil, but we got back %v meaning the delete failed", stream)
    }
}
