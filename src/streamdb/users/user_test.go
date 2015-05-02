package users

import (
	"testing"
	"reflect"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)


func TestCreateUser(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}
		err := testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email", "TestCreateUser_pass")
		require.Nil(t, err)

		err = testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email2", "TestCreateUser_pass2")
		require.NotNil(t, err)

		err = testdb.CreateUser("TestCreateUser_name2", "TestCreateUser_email", "TestCreateUser_pass2")
		require.NotNil(t, err)
	}
}
/**
func BenchmarkCreateUser(b *testing.B) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

	    for i := 0; i < b.N; i++ {
	        testdb.CreateUser(GetNextName(), GetNextEmail(), "TestCreateUser_pass2")
	    }
	}
}**/


func TestReadAllUsers(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		for i := 0; i < 5; i++{
			_, err := CreateTestUser(testdb)
			require.Nil(t, err)
		}

		users, err := testdb.ReadAllUsers()
		require.Nil(t, err)
		require.NotNil(t, users)


		err = testdb.CreateUser("TestReadAllUsers", "TestReadAllUsers_email", "TestReadAllUsers_pass")
		require.Nil(t, err)

		users2, err := testdb.ReadAllUsers()
		assert.Nil(t, err, "got err from read all users %v", err)
		require.NotNil(t, users2, "Could not get all users, was nil")

		assert.Equal(t, 1, len(users2) - len(users), "not taking into account changes")
	}
}

func TestReadUserByEmail(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		// test failures on non existance
		usr, err := testdb.ReadUserByEmail("doesnotexist   because spaces")
		assert.NotNil(t, err, "no error returned, expected non nil on failing case")

		// setup for reading
		err = testdb.CreateUser("TestReadUserByEmail_name", "TestReadUserByEmail_email", "TestReadUserByEmail_pass")
		require.Nil(t, err, "Could not create user for test reading...")

		usr, err = testdb.ReadUserByEmail("TestReadUserByEmail_email")
		assert.NotNil(t, usr, "did not get a user by email")
		assert.Nil(t, err, "got an error when trying to get a user that should exist %v", err)
	}
}

func TestReadUserByName(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		// test failures on non existance
		usr, err := testdb.ReadUserByName("")
		assert.NotNil(t, err)
		assert.NotNil(t, usr)

		// setup for reading
		err = testdb.CreateUser("TestReadUserByName_name", "TestReadUserByName_email", "TestReadUserByName_pass")
		require.Nil(t, err)

		usr, err = testdb.ReadUserByName("TestReadUserByName_name")
		assert.NotNil(t, usr, "did not get a user by name")
		assert.Nil(t, err, "got an error when trying to get a user that should exist")
	}
}

func TestReadUserById(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		// test failures on non existance
		usr, err := testdb.ReadUserById(-1)
		assert.NotNil(t, err)

		// setup for reading
		err = testdb.CreateUser("ReadUserById_name", "ReadUserById_email", "ReadUserById_pass")
		assert.Nil(t, err)

		usr, err = testdb.ReadUserByName("ReadUserById_name")
		assert.NotNil(t, usr)
		assert.Nil(t, err)
	}
}

func TestUpdateUser(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		err := testdb.UpdateUser(nil)
		assert.Equal(t, err, ERR_INVALID_PTR, "Didn't catch nil")

		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)


		usr.Name = "Hello"
		usr.Email = "hello@example.com"
		usr.Admin = true
		usr.UploadLimit_Items = 1
		usr.ProcessingLimit_S = 1
		usr.StorageLimit_Gb = 1

		err = testdb.UpdateUser(usr)
		require.Nil(t, err)

		usr2, err := testdb.ReadUserByName(usr.Name)

		if !reflect.DeepEqual(usr, usr2) {
			t.Errorf("The original and updated objects don't match orig: %v updated %v", usr, usr2)
		}
	}
}

func TestDeleteUser(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.DeleteUser(usr.UserId)
		require.Nil(t, err)

		_, err = testdb.ReadUserById(usr.UserId)
		require.NotNil(t, err, "The user with ID %v should have errored out, but it did not", usr.UserId)
	}
}

func TestReadStreamOwner(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		user, _, stream, err := CreateUDS(testdb)
		require.Nil(t, err)

		owner, err := testdb.ReadStreamOwner(stream.StreamId)
		require.Nil(t, err)

		require.Equal(t, owner.UserId, user.UserId, "Wrong stream owner got %v, expected %v", owner, user)
	}
}

func TestReadUserDevice(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		user, err := CreateTestUser(testdb)
		require.Nil(t, err)

		dev, err := testdb.ReadUserOperatingDevice(user)
		require.Nil(t, err)

		assert.Equal(t, dev.UserId, user.UserId, "Incorrect device returned.")

		user.Admin = true
		err = testdb.UpdateUser(user)
		require.Nil(t, err)

		dev, err = testdb.ReadUserOperatingDevice(user)
		require.Nil(t, err)

		assert.True(t, dev.IsAdmin)
	}
}

func TestLogin(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		user, err := CreateTestUser(testdb)

		_, _, err = testdb.Login(user.Name, testPassword)
		assert.Nil(t, err)

		_, _, err = testdb.Login(user.Email, testPassword)
		assert.Nil(t, err)

		_, _, err = testdb.Login("", testPassword)
		assert.Equal(t, err, InvalidUsernameError, "Wrong type returned %v", err)

		_, _, err = testdb.Login(user.Name, "")
		assert.NotNil(t, err, "Accepted blank password")
	}
}

func TestUpgradePassword(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		user, err := CreateTestUser(testdb)
		require.Nil(t, err)

		res := user.UpgradePassword(testPassword)
		assert.Equal(t, res, false, "Should not need to upgrade a password with the same scheme")

		user.PasswordHashScheme = ""
		res = user.UpgradePassword(testPassword)
		assert.Equal(t, res, true, "Should want to upgrade a password with an old has type")

		assert.NotEqual(t, "", user.PasswordHashScheme, "The has scheme was not updated")
	}
}

func TestRevertUneditableFields(t *testing.T) {
	orig := User{1,"Name", "Email", "Password", "passsalt", "hash", true, 1,1,1}
	root := User{1,"", "", "", "", "", false, 0,0,0}
	user := User{1,"Name", "", "", "", "", true, 1,1,1}

	// the one we're going to try to submit
	blank := User{0, "", "", "", "", "", false, 0,0,0}

	tmpu := orig
	tmpu.RevertUneditableFields(blank, ROOT)
	assert.Equal(t, tmpu, root, "Conversion as root didn't work got %v, expected %v", tmpu, root)

	tmpu = orig
	tmpu.RevertUneditableFields(blank, USER)
	assert.Equal(t, tmpu , user, "Conversion as user didn't work got %v, expected %v", tmpu, root)
}
