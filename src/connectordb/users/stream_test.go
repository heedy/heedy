/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	streamtestType = "{\"type\":\"string\"}"
)

func TestCreateStream(t *testing.T) {
	for _, testdb := range testdatabases {
		_, dev, stream, err := CreateUDS(testdb)
		require.Nil(t, err)

		err = testdb.CreateStream(stream.Name, streamtestType, dev.DeviceId)
		assert.NotNil(t, err, "Created stream with duplicate name")

		// Test with invalid schema
		err = testdb.CreateStream("tcs_001", "{", dev.DeviceId)
		assert.NotNil(t, err, "Created stream with invalid schema")

		// Test with embedded objects
		/*
			err = testdb.CreateStream("tcs_002", `{
			"type":"object",
			"properties":{
					"foo":{
						"type":"object"
					}
				}
			}`, dev.DeviceId)
			assert.NotNil(t, err, "Created stream with object schema")
		*/
	}
}

func TestUpdateStream(t *testing.T) {

	for _, testdb := range testdatabases {
		_, _, stream, err := CreateUDS(testdb)
		require.Nil(t, err)

		stream.Nickname = "true"
		stream.Type = streamtestType

		err = testdb.UpdateStream(stream)
		assert.Nil(t, err, "Could not update stream %v", err)

		stream2, err := testdb.ReadStreamById(stream.StreamId)
		require.Nil(t, err, "got an error when trying to get a stream that should exist %v", err)

		if !reflect.DeepEqual(stream, stream2) {
			t.Errorf("The original and updated objects don't match orig: %v updated %v", stream, stream2)
		}

		err = testdb.UpdateStream(nil)
		assert.Equal(t, err, InvalidPointerError, "Function safeguards failed")
	}
}

func TestDeleteStream(t *testing.T) {

	for _, testdb := range testdatabases {
		_, _, stream, err := CreateUDS(testdb)
		require.Nil(t, err)

		err = testdb.DeleteStream(stream.StreamId)
		require.Nil(t, err, "Error when attempted delete %v", err)

		_, err = testdb.ReadStreamById(stream.StreamId)
		require.NotNil(t, err, "The stream with the selected ID should have errored out, but it was not")
	}
}

func TestReadStreamByDevice(t *testing.T) {

	for _, testdb := range testdatabases {
		_, dev, _, err := CreateUDS(testdb)
		require.Nil(t, err)

		testdb.CreateStream("TestReadStreamByDevice2", streamtestType, dev.DeviceId)

		streams, err := testdb.ReadStreamsByDevice(dev.DeviceId)
		require.Nil(t, err)
		require.Len(t, streams, 2, "didn't get enough streams")
	}
}

func TestReadStreamsByUser(t *testing.T) {
	for _, testdb := range testdatabases {
		inserted := map[Stream]bool{}

		user, _, stream, err := CreateUDS(testdb)
		require.Nil(t, err)
		require.NotNil(t, user)
		require.NotNil(t, stream)

		fmt.Printf("User Id: %v\n", user.UserId)

		inserted[*stream] = true

		// create a bunch of devices
		for i := 0; i < 10; i++ {
			device, err := CreateTestDevice(testdb, user)

			require.Nil(t, err)

			fmt.Printf("Device Id: %v\n", device.DeviceId)

			// create a bunch of streams
			for j := 0; j < 10; j++ {
				stream, err := CreateTestStream(testdb, device)
				require.Nil(t, err)
				inserted[*stream] = true
			}
		}

		// Test selecting them
		streams, err := testdb.ReadStreamsByUser(user.UserId)
		require.Nil(t, err, "Retrieved streams was nil")

		// We need to add in the other missing log stream.
		require.Equal(t, len(inserted)+1, len(streams), "Wrong number of streams returned")
	}
}
