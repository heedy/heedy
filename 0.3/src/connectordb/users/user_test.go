/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package users

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {

	for _, testdb := range testdatabases {
		err := testdb.CreateUser(&UserMaker{User: User{Name: "TestCreateUser_name", Email: "TestCreateUser_email", Password: "TestCreateUser_pass", Role: "test"}})
		require.Nil(t, err)

		err = testdb.CreateUser(&UserMaker{User: User{Name: "TestCreateUser_name", Email: "TestCreateUser_email2", Password: "TestCreateUser_pass2", Role: "test"}})
		require.NotNil(t, err)

		err = testdb.CreateUser(&UserMaker{User: User{Name: "TestCreateUser_name2", Email: "TestCreateUser_email", Password: "TestCreateUser_pass2", Role: "test"}})
		require.NotNil(t, err)
	}
}

func TestReadAllUsers(t *testing.T) {

	for _, testdb := range testdatabases {
		for i := 0; i < 5; i++ {
			_, err := CreateTestUser(testdb)
			require.Nil(t, err)
		}

		users, err := testdb.ReadAllUsers()
		require.Nil(t, err)
		require.NotNil(t, users)

		err = testdb.CreateUser(&UserMaker{User: User{Name: "TestReadAllUsers", Email: "TestReadAllUsers_email", Password: "TestReadAllUsers_pass", Role: "test"}})
		require.Nil(t, err)

		users2, err := testdb.ReadAllUsers()
		assert.Nil(t, err, "got err from read all users %v", err)
		require.NotNil(t, users2, "Could not get all users, was nil")

		assert.Equal(t, 1, len(users2)-len(users), "not taking into account changes")
	}
}

func TestReadUserByName(t *testing.T) {

	for _, testdb := range testdatabases {
		// test failures on non existance
		usr, err := testdb.ReadUserByName("")
		assert.NotNil(t, err)

		// setup for reading
		err = testdb.CreateUser(&UserMaker{User: User{Name: "TestReadUserByName_name", Email: "TestReadUserByName_email", Password: "TestReadUserByName_pass", Role: "test"}})
		require.Nil(t, err)

		usr, err = testdb.ReadUserByName("TestReadUserByName_name")
		assert.NotNil(t, usr, "did not get a user by name")
		assert.Nil(t, err, "got an error when trying to get a user that should exist")
	}
}

func TestReadUserById(t *testing.T) {

	for _, testdb := range testdatabases {
		// test failures on non existance
		usr, err := testdb.ReadUserById(-1)
		assert.NotNil(t, err)

		// setup for reading
		err = testdb.CreateUser(&UserMaker{User: User{Name: "ReadUserById_name", Email: "ReadUserById_email", Password: "ReadUserById_pass", Role: "test"}})
		assert.Nil(t, err)

		usr, err = testdb.ReadUserByName("ReadUserById_name")
		assert.NotNil(t, usr)
		assert.Nil(t, err)
	}
}

func TestUpdateUser(t *testing.T) {

	for _, testdb := range testdatabases {
		err := testdb.UpdateUser(nil)
		assert.Equal(t, err, InvalidPointerError, "Didn't catch nil")

		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		usr.Name = "Hello"
		usr.Email = "hello@example.com"

		err = testdb.UpdateUser(usr)
		require.Nil(t, err)

		usr2, err := testdb.ReadUserByName(usr.Name)

		if !reflect.DeepEqual(usr, usr2) {
			t.Errorf("The original and updated objects don't match orig: %v updated %v", usr, usr2)
		}
	}
}

func TestDeleteUser(t *testing.T) {

	for _, testdb := range testdatabases {
		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.DeleteUser(usr.UserID)
		require.Nil(t, err)

		_, err = testdb.ReadUserById(usr.UserID)
		require.NotNil(t, err, "The user with ID %v should have errored out, but it did not", usr.UserID)

		err = testdb.DeleteUser(usr.UserID)
		require.Equal(t, err, ErrNothingToDelete, "Didn't catch try to delete deleted user")
	}
}

func TestReadUserDevice(t *testing.T) {

	for _, testdb := range testdatabases {
		user, err := CreateTestUser(testdb)
		require.Nil(t, err)

		dev, err := testdb.ReadUserOperatingDevice(user)
		require.Nil(t, err)

		assert.Equal(t, dev.UserID, user.UserID, "Incorrect device returned.")

		user.Role = "test2"
		err = testdb.UpdateUser(user)
		require.Nil(t, err)
	}
}

func TestLogin(t *testing.T) {

	for _, testdb := range testdatabases {
		user, err := CreateTestUser(testdb)

		_, _, err = testdb.Login(user.Name, testPassword)
		assert.Nil(t, err)

		_, _, err = testdb.Login(user.Email, testPassword)
		assert.Nil(t, err)

		_, _, err = testdb.Login("", testPassword)
		assert.Equal(t, err, ErrLoginFailed, "Wrong type returned %v", err)

		_, _, err = testdb.Login(user.Name, "")
		assert.NotNil(t, err, "Accepted blank password")
	}
}

func TestUpgradePassword(t *testing.T) {

	for _, testdb := range testdatabases {
		user, err := CreateTestUser(testdb)
		require.Nil(t, err)

		res := user.UpgradePassword(testPassword)
		assert.Equal(t, res, false, "Should not need to upgrade a password with the same scheme")

		user.PasswordHashScheme = ""
		res = user.UpgradePassword(testPassword)
		assert.Equal(t, res, true, "Should want to upgrade a password with an old hash type")

		assert.NotEqual(t, "", user.PasswordHashScheme, "The hash scheme was not updated")
	}
}
