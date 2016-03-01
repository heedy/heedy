/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.

This file provides the initialization of the test procedures

**/
package users

import (
	"config"
	"dbsetup/dbutil"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

var (
	nextNameId  = 0
	nextEmailId = 0

	testPostgres       UserDatabase
	testdatabases      = []UserDatabase{}
	testdatabasesNames = []string{}
	testPassword       = "P@$$W0Rd123"
)

func GetNextName() string {
	nextNameId++
	return "name_" + strconv.Itoa(nextNameId)
}

func GetNextEmail() string {
	nextEmailId++
	return "name" + strconv.Itoa(nextNameId) + "@domain.com"
}

func init() {
	testPostgres := initDB(config.TestOptions.SqlConnectionString)
	testdatabases = []UserDatabase{testPostgres}
	testdatabasesNames = []string{"postgres"}
}

func initDB(dbName string) UserDatabase {
	_ = os.Remove(dbName) // may fail if postgres

	// Init the db
	err := dbutil.UpgradeDatabase(dbName, true)
	if err != nil {
		panic(err.Error())
	}

	sql, dbtype, err := dbutil.OpenSqlDatabase(dbName)
	if err != nil {
		log.Panic(err)
	}

	db := NewUserDatabase(sql, dbtype, false, 1, 1, 1)

	return db
}

func CreateTestStream(testdb UserDatabase, dev *Device) (*Stream, error) {
	name := GetNextName()
	err := testdb.CreateStream(&StreamMaker{Stream: Stream{Name: name, Schema: "{\"type\":\"number\"}", DeviceID: dev.DeviceID}})
	if err != nil {
		return nil, err
	}

	return testdb.ReadStreamByDeviceIDAndName(dev.DeviceID, name)
}

func CreateTestUser(testdb UserDatabase) (*User, error) {
	name := GetNextName()
	email := GetNextEmail()

	//log.Printf("Creating test user with name: %v, email: %v, pass: %v", name, email, testPassword)

	err := testdb.CreateUser(&UserMaker{User: User{Name: name, Email: email, Password: testPassword, Role: "test"}})

	if err != nil {
		return nil, err
	}

	return testdb.ReadUserByName(name)
}

func CreateTestDevice(testdb UserDatabase, usr *User) (*Device, error) {
	name := GetNextName()
	err := testdb.CreateDevice(&DeviceMaker{Device: Device{Name: name, UserID: usr.UserID}})
	if err != nil {
		return nil, err
	}

	return testdb.ReadDeviceForUserByName(usr.UserID, name)
}

// Creates a connected user, device and stream
func CreateUDS(testdb UserDatabase) (*User, *Device, *Stream, error) {
	u, err := CreateTestUser(testdb)

	if err != nil {
		return nil, nil, nil, err
	}

	d, err := CreateTestDevice(testdb, u)
	if err != nil {
		return nil, nil, nil, err
	}
	s, err := CreateTestStream(testdb, d)

	return u, d, s, err
}
