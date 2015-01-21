package users

import (
    "testing"
    "reflect"
    "log"
    )

var (
    usrid2 int64
    usr2 *User
)


func init() {
    var err error

    usrid, err = CreateUser("DeviceTestUserName", "DeviceTestUserEmail", "DeviceTestUserPassword")
    if err != nil {
        log.Print(err)
    }

    usr, err = ReadUserById(usrid)
    if err != nil {
        log.Print(err)
    }

    usrid2, err = CreateUser("DeviceTestUserName2", "DeviceTestUserEmail2", "DeviceTestUserPassword2")
    if err != nil {
        log.Print(err)
    }

    usr2, err = ReadUserById(usrid2)
    if err != nil {
        log.Print(err)
    }

}


func TestCreateDevice(t *testing.T) {
    _, err := CreateDevice("TestCreateDevice", usr)
    if(err != nil) {
        t.Errorf("Cannot create device %v", err)
        return
    }

    // DeviceName/Usernames must be unique
    _, err = CreateDevice("TestCreateDevice", usr)
    if(err == nil) {
        t.Errorf("Created device with duplicate name under the same user")
    }

    // but should work with different users
    _, err = CreateDevice("TestCreateDevice", usr2)
    if(err != nil) {
        t.Errorf("Could not create device with secnod user %v", err)
        return
    }
}


func TestReadDeviceById(t *testing.T) {
    id, err := CreateDevice("TestReadStreamById", usr)

    if(err != nil) {
        t.Errorf("Cannot create object %v", err)
        return
    }

    obj, err := ReadDeviceById(id)

    if err != nil || obj == nil {
        t.Errorf("Cannot read object back with returned id %v, err: %v, obj: %v", id, err, obj)
        return
    }
}

func TestUpdateDevice(t *testing.T) {
    id, err := CreateDevice("TestUpdateDevice", usr)
    if err != nil {
        t.Errorf("Cannot create object %v", err)
        return
    }

    obj, err := ReadDeviceById(id)
    if err != nil || obj == nil {
        t.Errorf("Cannot read object back with id: {}, err: {}, obj:{}", id, err, obj)
        return
    }

    obj.Name = "Test"
    obj.ApiKey = obj.ApiKey + "Testing" // should work with all UUIDs, still will be unique
    obj.Enabled = false
    obj.Icon_PngB64 = ""
    obj.Shortname = "My Wifi Router"
    obj.Superdevice = true

    err = UpdateDevice(obj)
    if err != nil {
        t.Errorf("Could not update object %v", err)
        return
    }

    obj2, err := ReadDeviceById(id)
    if err != nil || obj2 == nil {
        t.Errorf("Cannot read object back with id: {}, err: {}, obj:{}", id, err, obj2)
        return
    }

    if ! reflect.DeepEqual(obj, obj2) {
        t.Errorf("The original and updated objects don't match orig: %v updated %v", obj, obj2)
    }
}


func TestDeleteDevice(t *testing.T) {
    id, err := CreateDevice("TestDeleteDevice", usr)

    if nil != err {
        t.Errorf("Cannot create object to test delete err: %v", err)
        return
    }

    err = DeleteDevice(id)

    if nil != err {
        t.Errorf("Error when attempted delete %v", err)
        return
    }

    obj, err := ReadDeviceById(id)

    if err == nil {
        t.Errorf("The object with the selected ID should have errored out, but it did not")
        return
    }

    if obj != nil {
        t.Errorf("Expected nil, but we got back %v meaning the delete failed", obj)
    }
}
