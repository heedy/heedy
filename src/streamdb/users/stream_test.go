package users

import (
	"testing"
	"reflect"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

)



func TestCreateStream(t *testing.T) {

	CleanTestDB(testdb)
	_, dev, stream, err := CreateUDS(testdb)
	require.Nil(t, err)

	err = testdb.CreateStream(stream.Name, "{}", dev.DeviceId)
	assert.NotNil(t, err, "Created stream with duplicate name")
}


func TestUpdateStream(t *testing.T) {
	_, _, stream, err := CreateUDS(testdb)
	require.Nil(t, err)

	stream.Nickname = "true"
	stream.Type = "{a:'string'}"

	err = testdb.UpdateStream(stream)
	assert.Nil(t, err, "Could not update stream %v", err)

	stream2, err := testdb.ReadStreamById(stream.StreamId)
	require.Nil(t, err, "got an error when trying to get a stream that should exist %v", err)

	if !reflect.DeepEqual(stream, stream2) {
		t.Errorf("The original and updated objects don't match orig: %v updated %v", stream, stream2)
	}

	err = testdb.UpdateStream(nil)
	assert.Equal(t, err,  ERR_INVALID_PTR, "Function safeguards failed")
}

func TestDeleteStream(t *testing.T) {
	_, _, stream, err := CreateUDS(testdb)
	require.Nil(t, err)

	err = testdb.DeleteStream(stream.StreamId)
	require.Nil(t, err, "Error when attempted delete %v", err)

	_, err = testdb.ReadStreamById(stream.StreamId)
	require.NotNil(t, err, "The stream with the selected ID should have errored out, but it was not")
}

func TestReadStreamByDevice(t *testing.T) {
	_, dev, _, err := CreateUDS(testdb)
	require.Nil(t, err)

	testdb.CreateStream("TestReadStreamByDevice2", "{}", dev.DeviceId)

	streams, err := testdb.ReadStreamsByDevice(dev.DeviceId)
	require.Nil(t, err)
	require.Len(t, streams, 2, "didn't get enough streams")
}
