/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.

This file provides the initialization of the test procedures

**/
package users

import (
	"config"
	"dbsetup/dbutil"
	"strconv"
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
	testPostgres := initDB("postgres", config.TestOptions.SQLURI)
	testSqlite := initDB("sqlite3", "test.db")
	testdatabases = []UserDatabase{testPostgres, testSqlite}
	testdatabasesNames = []string{"postgres", "sqlite"}
}

func initDB(name, uri string) UserDatabase {
	err := dbutil.ClearDatabase(name, uri)
	if err != nil {
		if err.Error() != "remove test.db: no such file or directory" {
			panic(err.Error())
		}

	}

	err = dbutil.SetupDatabase(name, uri)
	if err != nil {
		panic(err.Error())
	}
	sql, err := dbutil.OpenDatabase(name, uri)
	if err != nil {
		panic(err.Error())
	}

	db := NewUserDatabase(sql, false, 1, 1, 1)

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
