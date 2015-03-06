package users

import (
    "testing"
    "reflect"
    "log"
    "os"
)

var (
    carid int64
    usrid int64
    usr *User
    devid int64
    dev *Device
    testdb *UserDatabase
    testdbname = "testing.sqlite3"
    usrid2 int64
    usr2 *User
)


func init() {
    var err error

    _ = os.Remove(testdbname)
    testdb, err = NewSqliteUserDatabase(testdbname)
    if err != nil {
        log.Print(err)
    }

    carid, err = testdb.CreatePhoneCarrier("StreamTestPhoneCarrier", "StreamTestPhoneCarrier.com")
    if err != nil {
        log.Print(err)
    }

    usrid, err = testdb.CreateUser("StreamTestUserName", "StreamTestUserEmail", "StreamTestUserPassword")
    if err != nil {
        log.Print(err)
    }

    usr, err = testdb.ReadUserById(usrid)
    if err != nil {
        log.Print(err)
    }

    devid, err = testdb.CreateDevice("StreamTestDevice", usr)
    if err != nil {
        log.Print(err)
    }

    dev, err = testdb.ReadDeviceById(devid)
    if err != nil {
        log.Print(err)
    }

    usrid, err = testdb.CreateUser("DeviceTestUserName", "DeviceTestUserEmail", "DeviceTestUserPassword")
    if err != nil {
        log.Print(err)
    }

    usr, err = testdb.ReadUserById(usrid)
    if err != nil {
        log.Print(err)
    }

    usrid2, err = testdb.CreateUser("DeviceTestUserName2", "DeviceTestUserEmail2", "DeviceTestUserPassword2")
    if err != nil {
        log.Print(err)
    }

    usr2, err = testdb.ReadUserById(usrid2)
    if err != nil {
        log.Print(err)
    }


}


func TestCreateStream(t *testing.T) {
    _, err := testdb.CreateStream("TestCreateStream", "{}" , dev)
    if(err != nil) {
        t.Errorf("Cannot create stream %v", err)
        return
    }

    _, err = testdb.CreateStream("TestCreateStream", "{}", dev)
    if(err == nil) {
        t.Errorf("Created stream with duplicate name")
    }

    _, err = testdb.CreateStream("", "", nil)
    if err != ERR_INVALID_PTR {
        t.Errorf("Function safeguards failed")
    }
}

func TestConstructStreamsFromRows(t *testing.T) {
    _, err := constructStreamsFromRows(nil)

    if err != ERR_INVALID_PTR {
        t.Errorf("Function safeguards failed")
    }

}

func TestReadStreamById(t *testing.T) {
    streamid, err := testdb.CreateStream("TestReadStreamById", "{}", dev)
    if(err != nil) {
        t.Errorf("Cannot create stream %v", err)
        return
    }

    stream, err := testdb.ReadStreamById(streamid)

    if err != nil || stream == nil {
        t.Errorf("Cannot read stream back with returned id %v", streamid)
        return
    }
}

func TestUpdateStream(t *testing.T) {
    streamid, err := testdb.CreateStream("TestUpdateStream", "{}", dev)
    if(err != nil) {
        t.Errorf("Cannot create stream %v", err)
        return
    }

    stream, err := testdb.ReadStreamById(streamid)

    if err != nil || stream == nil{
        t.Errorf("Cannot read stream back with returned id %v", streamid)
        return
    }

    stream.Name = "A"
    stream.Active = false
    stream.Public = true
    stream.Type = "{a:'string'}"
    //stream.OwnerId = dev

    err = testdb.UpdateStream(stream)

    if err != nil {
        t.Errorf("Could not update stream %v", err)
    }

    stream2, err := testdb.ReadStreamById(streamid)

    if err != nil {
        t.Errorf("got an error when trying to get a stream that should exist %v", err)
        return
    }

    if ! reflect.DeepEqual(stream, stream2) {
        t.Errorf("The original and updated objects don't match orig: %v updated %v", stream, stream2)
    }

    err = testdb.UpdateStream(nil)
    if err != ERR_INVALID_PTR {
        t.Errorf("Function safeguards failed")
    }

}

func TestDeleteStream(t *testing.T) {
    id, err := testdb.CreateStream("TestDeleteStream", "{}", dev)

    if nil != err {
        t.Errorf("Cannot create stream to test delete")
        return
    }

    err = testdb.DeleteStream(id)

    if nil != err {
        t.Errorf("Error when attempted delete %v", err)
        return
    }

    stream, err := testdb.ReadStreamById(id)

    if err == nil {
        t.Errorf("The stream with the selected ID should have errored out, but it was not")
        return
    }

    if stream != nil {
        t.Errorf("Expected nil, but we got back %v meaning the delete failed", stream)
    }
}

func TestReadStreamByDevice(t *testing.T) {
    testdb.CreateStream("TestReadStreamByDevice", "{}", dev)
    testdb.CreateStream("TestReadStreamByDevice2", "{}", dev)

    streams, err := testdb.ReadStreamsByDevice(dev)

    if err != nil {
        t.Errorf("Got error while reading streams by device")
        return
    }

    // TODO change this to look for proper streams once we have a set test db
    // with fixed items
    if len(streams) < 2 {
        t.Errorf("didn't get enough streams")
    }

    _, err = testdb.ReadStreamsByDevice(nil)

    if err != ERR_INVALID_PTR {
        t.Errorf("Function safeguards failed")
    }
}
