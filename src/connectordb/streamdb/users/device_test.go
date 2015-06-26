package users

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDevice(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		usr2, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.CreateDevice("TestCreateDevice", usr.UserId)
		require.Nil(t, err)

		// DeviceName/Usernames must be unique
		err = testdb.CreateDevice("TestCreateDevice", usr.UserId)
		assert.NotNil(t, err, "Created device with duplicate name under the same user")

		// but should work with different users
		err = testdb.CreateDevice("TestCreateDevice", usr2.UserId)
		assert.Nil(t, err, "Could not create device with secnod user %v", err)
	}
}

func TestReadDeviceById(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.CreateDevice("TestReadStreamById", usr.UserId)
		require.Nil(t, err)

		devforid, err := testdb.ReadDeviceForUserByName(usr.UserId, "TestReadStreamById")
		require.Nil(t, err)

		id := devforid.DeviceId
		obj, err := testdb.ReadDeviceById(id)
		require.Nil(t, err)
		require.NotNil(t, obj)
	}
}

func TestReadDeviceByApiKey(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		_, dev, _, err := CreateUDS(testdb)
		require.Nil(t, err)

		key := dev.ApiKey

		dev2, err := testdb.ReadDeviceByApiKey(key)
		require.Nil(t, err)
		assert.Equal(t, key, dev2.ApiKey)

		dev3, err := testdb.ReadDeviceByApiKey("")
		assert.NotNil(t, err, "non existing device read by api key, dev %v", dev3)
	}

}

func TestUpdateDevice(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.CreateDevice("TestUpdateDevice", usr.UserId)
		require.Nil(t, err)

		obj, err := testdb.ReadDeviceForUserByName(usr.UserId, "TestUpdateDevice")
		require.Nil(t, err)
		require.NotNil(t, obj)

		obj.ApiKey = obj.ApiKey + "Testing" // should work with all UUIDs, still will be unique
		obj.Enabled = false
		obj.Nickname = "My Wifi Router"
		obj.IsAdmin = true

		err = testdb.UpdateDevice(obj)
		require.Nil(t, err)

		obj2, err := testdb.ReadDeviceForUserByName(usr.UserId, "TestUpdateDevice")
		require.Nil(t, err)
		require.NotNil(t, obj2)

		if !reflect.DeepEqual(obj, obj2) {
			t.Errorf("The original and updated objects don't match orig: %v updated %v", obj, obj2)
		}

		err = testdb.UpdateDevice(nil)
		assert.Equal(t, err, ERR_INVALID_PTR, "Allowed nil pointer through")
	}
}

func TestDeleteDevice(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		usr, obj, _, err := CreateUDS(testdb)
		require.Nil(t, err)

		err = testdb.DeleteDevice(obj.DeviceId)
		assert.Nil(t, err, "error on delete %v", err)

		obj, err = testdb.ReadDeviceForUserByName(usr.UserId, "TestDeleteDevice")
		assert.NotNil(t, err, "should not succeed, device should be gone %v", err)
	}
}
