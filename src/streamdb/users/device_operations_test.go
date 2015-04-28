package users

import (
	"reflect"
	"testing"
)

func TestCreateDevice(t *testing.T) {
	usr, err := CreateTestUser(testdb)
	if err != nil {
		t.Errorf("Cannot create test user %v", err.Error())
		return
	}

	usr2, err := CreateTestUser(testdb)
	if err != nil {
		t.Errorf("Cannot create test user %v", err.Error())
		return
	}


	err = testdb.CreateDevice("TestCreateDevice", usr.UserId)
	if err != nil {
		t.Errorf("Cannot create device %v", err)
		return
	}

	// DeviceName/Usernames must be unique
 	err = testdb.CreateDevice("TestCreateDevice", usr.UserId)
	if err == nil {
		t.Errorf("Created device with duplicate name under the same user")
	}

	// but should work with different users
	err = testdb.CreateDevice("TestCreateDevice", usr2.UserId)
	if err != nil {
		t.Errorf("Could not create device with secnod user %v", err)
		return
	}
}

func TestReadDeviceById(t *testing.T) {
	usr, err := CreateTestUser(testdb)
	if err != nil {
		t.Errorf("Cannot create test user %v", err.Error())
		return
	}

	err = testdb.CreateDevice("TestReadStreamById", usr.UserId)
	if err != nil {
		t.Errorf("Cannot create object %v", err)
		return
	}

	devforid, err := testdb.ReadDeviceForUserByName(usr.UserId, "TestReadStreamById")
	if err != nil {
		t.Errorf("Cannot get id %v", err)
		return
	}
	id := devforid.UserId

	obj, err := testdb.ReadDeviceById(id)

	if err != nil || obj == nil {
		t.Errorf("Cannot read object back with returned id %v, err: %v, obj: %v", id, err, obj)
		return
	}
}

func TestUpdateDevice(t *testing.T) {
	usr, err := CreateTestUser(testdb)
	if err != nil {
		t.Errorf("Cannot create test user %v", err.Error())
		return
	}

	err = testdb.CreateDevice("TestUpdateDevice", usr.UserId)
	if err != nil {
		t.Errorf("Cannot create object %v", err)
		return
	}

	obj, err := testdb.ReadDeviceForUserByName(usr.UserId, "TestUpdateDevice")
	if err != nil || obj == nil {
		t.Errorf("Cannot read object back with err: %v, obj:%v", err, obj)
		return
	}

	obj.ApiKey = obj.ApiKey + "Testing" // should work with all UUIDs, still will be unique
	obj.Enabled = false
	obj.Nickname = "My Wifi Router"
	obj.IsAdmin = true

	err = testdb.UpdateDevice(obj)
	if err != nil {
		t.Errorf("Could not update object %v", err)
		return
	}

	obj2, err := testdb.ReadDeviceForUserByName(usr.UserId, "TestUpdateDevice")
	if err != nil || obj2 == nil {
		t.Errorf("Cannot read object back with err: {}, obj:{}", err, obj2)
		return
	}

	if !reflect.DeepEqual(obj, obj2) {
		t.Errorf("The original and updated objects don't match orig: %v updated %v", obj, obj2)
	}

	err = testdb.UpdateDevice(nil)
	if err != ERR_INVALID_PTR {
		t.Errorf("Allowed nil pointer through")
	}
}

func TestDeleteDevice(t *testing.T) {
	usr, err := CreateTestUser(testdb)
	if err != nil {
		t.Errorf("Cannot create test user %v", err.Error())
		return
	}

	err = testdb.CreateDevice("TestDeleteDevice", usr.UserId)

	if nil != err {
		t.Errorf("Cannot create object to test delete err: %v", err)
		return
	}

	obj, err := testdb.ReadDeviceForUserByName(usr.UserId, "TestDeleteDevice")
	if nil != err {
		t.Errorf("Cannot create object to test delete err: %v", err)
		return
	}

	err = testdb.DeleteDevice(obj.DeviceId)

	if nil != err {
		t.Errorf("Error when attempted delete %v", err)
		return
	}

	obj, err = testdb.ReadDeviceForUserByName(usr.UserId, "TestDeleteDevice")

	if err == nil {
		t.Errorf("The object with the selected ID should have errored out, but it did not")
		return
	}
}
