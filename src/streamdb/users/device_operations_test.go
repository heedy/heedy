package users

import (
    "testing"
    "reflect"
    )


func TestCreateDevice(t *testing.T) {
    _, err := testdb.CreateDevice("TestCreateDevice", usr)
    if(err != nil) {
        t.Errorf("Cannot create device %v", err)
        return
    }

    // DeviceName/Usernames must be unique
    _, err = testdb.CreateDevice("TestCreateDevice", usr)
    if(err == nil) {
        t.Errorf("Created device with duplicate name under the same user")
    }

    // but should work with different users
    _, err = testdb.CreateDevice("TestCreateDevice", usr2)
    if(err != nil) {
        t.Errorf("Could not create device with secnod user %v", err)
        return
    }

    _, err = testdb.CreateDevice("TestCreateDevice2", nil)
    if err != ERR_INVALID_PTR {
        t.Errorf("allowed user to be nil")
    }
}


func TestReadDeviceById(t *testing.T) {
    id, err := testdb.CreateDevice("TestReadStreamById", usr)

    if(err != nil) {
        t.Errorf("Cannot create object %v", err)
        return
    }

    obj, err := testdb.ReadDeviceById(id)

    if err != nil || obj == nil {
        t.Errorf("Cannot read object back with returned id %v, err: %v, obj: %v", id, err, obj)
        return
    }
}

func TestUpdateDevice(t *testing.T) {
    id, err := testdb.CreateDevice("TestUpdateDevice", usr)
    if err != nil {
        t.Errorf("Cannot create object %v", err)
        return
    }

    obj, err := testdb.ReadDeviceById(id)
    if err != nil || obj == nil {
        t.Errorf("Cannot read object back with id: %v, err: %v, obj:%v", id, err, obj)
        return
    }

    obj.Name = "Test"
    obj.ApiKey = obj.ApiKey + "Testing" // should work with all UUIDs, still will be unique
    obj.Enabled = false
    obj.Icon_PngB64 = ""
    obj.Shortname = "My Wifi Router"
    obj.Superdevice = true

    err = testdb.UpdateDevice(obj)
    if err != nil {
        t.Errorf("Could not update object %v", err)
        return
    }

    obj2, err := testdb.ReadDeviceById(id)
    if err != nil || obj2 == nil {
        t.Errorf("Cannot read object back with id: {}, err: {}, obj:{}", id, err, obj2)
        return
    }

    if ! reflect.DeepEqual(obj, obj2) {
        t.Errorf("The original and updated objects don't match orig: %v updated %v", obj, obj2)
    }

    err = testdb.UpdateDevice(nil)
    if err != ERR_INVALID_PTR {
        t.Errorf("Allowed nil pointer through")
    }
}

func TestConstructDevicesFromRows(t *testing.T) {
    _, err := constructDevicesFromRows(nil, nil)

    if err != ERR_INVALID_PTR {
        t.Errorf("Allowed nil pointer through")
    }

    _, err = constructDevicesFromRows(nil, ERR_INVALID_PTR)
    if err != ERR_INVALID_PTR {
        t.Errorf("Passed up error checking")
    }
}


func TestDeleteDevice(t *testing.T) {
    id, err := testdb.CreateDevice("TestDeleteDevice", usr)

    if nil != err {
        t.Errorf("Cannot create object to test delete err: %v", err)
        return
    }

    err = testdb.DeleteDevice(id)

    if nil != err {
        t.Errorf("Error when attempted delete %v", err)
        return
    }

    obj, err := testdb.ReadDeviceById(id)

    if err == nil {
        t.Errorf("The object with the selected ID should have errored out, but it did not")
        return
    }

    if obj != nil {
        t.Errorf("Expected nil, but we got back %v meaning the delete failed", obj)
    }
}
