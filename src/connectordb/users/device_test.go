/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDevice(t *testing.T) {
	for _, testdb := range testdatabases {
		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		usr2, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.CreateDevice("TestCreateDevice", usr.UserID, 0)
		require.Nil(t, err)

		// DeviceName/Usernames must be unique
		err = testdb.CreateDevice("TestCreateDevice", usr.UserID, 0)
		assert.NotNil(t, err, "Created device with duplicate name under the same user")

		// but should work with different users
		err = testdb.CreateDevice("TestCreateDevice", usr2.UserID, 0)
		assert.Nil(t, err, "Could not create device with secnod user %v", err)
	}
}

func TestReadDeviceByID(t *testing.T) {
	for _, testdb := range testdatabases {
		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.CreateDevice("TestReadStreamByID", usr.UserID, 0)
		require.Nil(t, err)

		devforid, err := testdb.ReadDeviceForUserByName(usr.UserID, "TestReadStreamByID")
		require.Nil(t, err)

		id := devforid.DeviceID
		obj, err := testdb.ReadDeviceByID(id)
		require.Nil(t, err)
		require.NotNil(t, obj)
	}
}

func TestReadDeviceByAPIKey(t *testing.T) {
	for _, testdb := range testdatabases {
		_, dev, _, err := CreateUDS(testdb)
		require.Nil(t, err)

		key := dev.APIKey

		dev2, err := testdb.ReadDeviceByAPIKey(key)
		require.Nil(t, err)
		assert.Equal(t, key, dev2.APIKey)

		dev3, err := testdb.ReadDeviceByAPIKey("")
		assert.NotNil(t, err, "non existing device read by api key, dev %v", dev3)
	}

}

func TestUpdateDevice(t *testing.T) {
	for _, testdb := range testdatabases {

		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.CreateDevice("TestUpdateDevice", usr.UserID, 0)
		require.Nil(t, err)

		obj, err := testdb.ReadDeviceForUserByName(usr.UserID, "TestUpdateDevice")
		require.Nil(t, err)
		require.NotNil(t, obj)

		obj.APIKey = obj.APIKey + "Testing" // should work with all UUIDs, still will be unique
		obj.Enabled = false
		obj.Nickname = "My Wifi Router"

		err = testdb.UpdateDevice(obj)
		require.Nil(t, err)

		obj2, err := testdb.ReadDeviceForUserByName(usr.UserID, "TestUpdateDevice")
		require.Nil(t, err)
		require.NotNil(t, obj2)

		if !reflect.DeepEqual(obj, obj2) {
			t.Errorf("The original and updated objects don't match orig: %v updated %v", obj, obj2)
		}

		err = testdb.UpdateDevice(nil)
		assert.Equal(t, err, InvalidPointerError, "Allowed nil pointer through")
	}
}

func TestDeleteDevice(t *testing.T) {
	for _, testdb := range testdatabases {
		usr, obj, _, err := CreateUDS(testdb)
		require.Nil(t, err)

		err = testdb.DeleteDevice(obj.DeviceID)
		assert.Nil(t, err, "error on delete %v", err)

		obj, err = testdb.ReadDeviceForUserByName(usr.UserID, "TestDeleteDevice")
		assert.NotNil(t, err, "should not succeed, device should be gone %v", err)
	}
}
