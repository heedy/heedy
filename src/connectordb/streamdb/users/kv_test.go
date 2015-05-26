package users

import (
	"testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)


func TestCreateKv(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}
		user, device, stream, _ := CreateUDS(testdb)

		// test original works
		err := testdb.CreateUserKeyValue(user.UserId, "testkey", "testvalue")
		assert.Nil(t, err)

		err = testdb.CreateDeviceKeyValue(device.DeviceId, "testkey", "testvalue")
		assert.Nil(t, err)

		err = testdb.CreateStreamKeyValue(stream.StreamId, "testkey", "testvalue")
		assert.Nil(t, err)

		// test duplicates fail
		err = testdb.CreateUserKeyValue(user.UserId, "testkey", "testvalue")
		assert.NotNil(t, err)

		err = testdb.CreateDeviceKeyValue(device.DeviceId, "testkey", "testvalue")
		assert.NotNil(t, err)

		err = testdb.CreateStreamKeyValue(stream.StreamId, "testkey", "testvalue")
		assert.NotNil(t, err)
		/**
		// test invalid ids
		err = testdb.CreateUserKeyValue(-1, "a", "testvalue")
		assert.NotNil(t, err)

		err = testdb.CreateDeviceKeyValue(-1, "a", "testvalue")
		assert.NotNil(t, err)

		err = testdb.CreateStreamKeyValue(-1, "a", "testvalue")
		assert.NotNil(t, err)
		**/
	}
}



func TestUpdateKv(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}
		user, device, stream, _ := CreateUDS(testdb)

		// do inserts
		err := testdb.CreateUserKeyValue(user.UserId, "testkey", "testvalue")
		require.Nil(t, err)

		err = testdb.CreateDeviceKeyValue(device.DeviceId, "testkey", "testvalue")
		require.Nil(t, err)

		err = testdb.CreateStreamKeyValue(stream.StreamId, "testkey", "testvalue")
		require.Nil(t, err)

		// test updates
		ukv := UserKeyValue{user.UserId, "testkey", "2"}
		err = testdb.UpdateUserKeyValue(ukv)
		assert.Nil(t, err)

		dkv := DeviceKeyValue{device.DeviceId, "testkey", "2"}
		err = testdb.UpdateDeviceKeyValue(dkv)
		assert.Nil(t, err)

		skv := StreamKeyValue{stream.StreamId, "testkey", "2"}
		err = testdb.UpdateStreamKeyValue(skv)
		assert.Nil(t, err)

		ukv2, err := testdb.ReadUserKeyValue(user.UserId, "testkey")
		assert.Nil(t, err)
		assert.Equal(t, "2", ukv2.Value)

		dkv2, err := testdb.ReadDeviceKeyValue(device.DeviceId, "testkey")
		assert.Nil(t, err)
		assert.Equal(t, "2", dkv2.Value)

		skv2, err := testdb.ReadStreamKeyValue(stream.StreamId, "testkey")
		assert.Nil(t, err)
		assert.Equal(t, "2", skv2.Value)
	}
}
