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
	testdatabases      = []SqlUserDatabase{}
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
	testdatabases = []SqlUserDatabase{testPostgres}
	testdatabasesNames = []string{"postgres"}
}

func initDB(dbName string) SqlUserDatabase {
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

	db := SqlUserDatabase{}
	db.initSqlUserDatabase(sql, dbtype)

	return db
}

func CreateTestStream(testdb SqlUserDatabase, dev *Device) (*Stream, error) {
	name := GetNextName()
	err := testdb.CreateStream(name, "{\"type\":\"number\"}", dev.DeviceId)
	if err != nil {
		return nil, err
	}

	return testdb.ReadStreamByDeviceIdAndName(dev.DeviceId, name)
}

func CreateTestUser(testdb SqlUserDatabase) (*User, error) {
	name := GetNextName()
	email := GetNextEmail()

	//log.Printf("Creating test user with name: %v, email: %v, pass: %v", name, email, testPassword)

	err := testdb.CreateUser(name, email, testPassword)

	if err != nil {
		return nil, err
	}

	return testdb.ReadUserByName(name)
}

func CreateTestDevice(testdb SqlUserDatabase, usr *User) (*Device, error) {
	name := GetNextName()
	err := testdb.CreateDevice(name, usr.UserId)
	if err != nil {
		return nil, err
	}

	return testdb.ReadDeviceForUserByName(usr.UserId, name)
}

// Creates a connected user, device and stream
func CreateUDS(testdb SqlUserDatabase) (*User, *Device, *Stream, error) {
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
