package streamdb

import (
    "streamdb/users"
    )

type Device struct {
    Db *Database
    Device *users.Device
}

//Returns the Administrator device (which has all possible permissions)
//Having a nil users.Device means that it is administrator
func (db *Database) GetAdminDevice() *Device {
    return &Device{db,nil}
}

//Given an API key, returns the  Device object
func (db *Database) GetDevice(apikey string) (*Device,error) {
    dev,err := db.ReadDeviceByApiKey(apikey)
    if err!=nil {
        return nil,err
    }
    return &Device{db,dev},nil
}
