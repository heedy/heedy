package users

/**
This file provides the initialization of the test procedures

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

import (
    "strconv"
    "os"
    "streamdb/dbutil"
    "log"
    )

var (
    nextNameId = 0
    nextEmailId = 0

	testSqlite3  *UserDatabase
    testPostgres *UserDatabase
    testdb *UserDatabase

	testdbname = "testing.sqlite3"
    testPassword = "P@$$W0Rd123"
)
/**
// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type UserTestSuite struct {
    suite.Suite
    testdb *UserDatabase
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *UserTestSuite) SetupTest() {
    suite.testdb = testSqlite3
    //CleanTestDB(testdb)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUsers(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}
**/

func GetNextName() string {
    nextNameId++
    return "name_" + strconv.Itoa(nextNameId)
}

func GetNextEmail() string {
    nextEmailId++
    return "name" + strconv.Itoa(nextNameId) + "@domain.com"
}




func init() {
	var err error

    // may not work if postgres
	_ = os.Remove(testdbname)

    // Init the db
    err = dbutil.UpgradeDatabase(testdbname, true)
    if err != nil {
        log.Panic("Could not set up db for testing: ", err.Error())
    }
	testSqlite3 = &UserDatabase{}

	sql, dbtype, err := dbutil.OpenSqlDatabase(testdbname)
	if err != nil {
		log.Panic(err)
	}

	testSqlite3.InitUserDatabase(sql, dbtype.String())

    testdb = testSqlite3

    CleanTestDB(testdb)
}




func CreateTestStream(testdb *UserDatabase, dev *Device) (*Stream, error) {
    name := GetNextName()
    err := testdb.CreateStream(name, "", dev.DeviceId)
    if err != nil {
        return nil, err
    }

    return testdb.ReadStreamByDeviceIdAndName(dev.DeviceId, name)
}


func CleanTestDB(testdb *UserDatabase){
    testdb.Exec("DELETE * FROM PhoneCarriers;")
    testdb.Exec("DELETE * FROM Users;")
    testdb.Exec("DELETE * FROM Devices;")
    testdb.Exec("DELETE * FROM Streams;")
    testdb.Exec("DELETE * FROM timeseriestable;")
    testdb.Exec("DELETE * FROM UserKeyValues;")
    testdb.Exec("DELETE * FROM DeviceKeyValues;")
    testdb.Exec("DELETE * FROM StreamKeyValues;")
}


func CreateTestUser(testdb *UserDatabase) (*User, error) {
    name := GetNextName()
    email := GetNextEmail()

    //log.Printf("Creating test user with name: %v, email: %v, pass: %v", name, email, testPassword)

    err := testdb.CreateUser(name, email, testPassword)

    if err != nil {
        return nil, err
    }

    return testdb.ReadUserByName(name)
}


func CreateTestDevice(testdb *UserDatabase, usr *User) (*Device, error) {
    name := GetNextName()
    err := testdb.CreateDevice(name, usr.UserId)
    if err != nil {
        return nil, err
    }

    return testdb.ReadDeviceForUserByName(usr.UserId, name)
}



// Creates a connected user, device and stream
func CreateUDS(testdb *UserDatabase) (*User, *Device, *Stream, error) {
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
