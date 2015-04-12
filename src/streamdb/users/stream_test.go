package users
/**

import (
	"testing"
)



func TestCreateStream(t *testing.T) {

	CleanTestDB()

	u, err := CreateTestUser()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	dev, err := CreateTestDevice(u)
	if err != nil {
		t.Errorf("%v\n", err)
	}


	err = testdb.CreateStream("TestCreateStream", "{}", dev.DeviceId)
	if err != nil {
		t.Errorf("Cannot create stream %v", err)
		return
	}

	err = testdb.CreateStream("TestCreateStream", "{}", dev.DeviceId)
	if err == nil {
		t.Errorf("Created stream with duplicate name")
	}
}


func TestUpdateStream(t *testing.T) {
	err := testdb.CreateStream("TestUpdateStream", "{}", dev.DeviceId)
	if err != nil {
		t.Errorf("Cannot create stream %v", err)
		return
	}

	stream, err := testdb.ReadStreamById(streamid)

	if err != nil || stream == nil {
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

	if !reflect.DeepEqual(stream, stream2) {
		t.Errorf("The original and updated objects don't match orig: %v updated %v", stream, stream2)
	}

	err = testdb.UpdateStream(nil)
	if err != ERR_INVALID_PTR {
		t.Errorf("Function safeguards failed")
	}

}**/
/**
func TestDeleteStream(t *testing.T) {
	id, err := testdb.CreateStream("TestDeleteStream", "{}", dev.DeviceId)

	if nil != err {
		t.Errorf("Cannot create stream to test delete")
		return
	}

	err = testdb.DeleteStream(id)

	if nil != err {
		t.Errorf("Error when attempted delete %v", err)
		return
	}

	stream, err := testdb.ReadStreamByName(id)

	if err == nil {
		t.Errorf("The stream with the selected ID should have errored out, but it was not")
		return
	}

	if stream != nil {
		t.Errorf("Expected nil, but we got back %v meaning the delete failed", stream)
	}
}

func TestReadStreamByDevice(t *testing.T) {
	testdb.CreateStream("TestReadStreamByDevice", "{}", dev.DeviceId)
	testdb.CreateStream("TestReadStreamByDevice2", "{}", dev.DeviceId)

	streams, err := testdb.ReadStreamsByDevice(dev.DeviceId)

	if err != nil {
		t.Errorf("Got error while reading streams by device")
		return
	}

	// TODO change this to look for proper streams once we have a set test db
	// with fixed items
	if len(streams) < 2 {
		t.Errorf("didn't get enough streams")
	}
}
**/
