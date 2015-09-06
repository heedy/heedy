package users

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {

	for _, testdb := range testdatabases {
		err := testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email", "TestCreateUser_pass")
		require.Nil(t, err)

		err = testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email2", "TestCreateUser_pass2")
		require.NotNil(t, err)

		err = testdb.CreateUser("TestCreateUser_name2", "TestCreateUser_email", "TestCreateUser_pass2")
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

		err = testdb.CreateUser("TestReadAllUsers", "TestReadAllUsers_email", "TestReadAllUsers_pass")
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
		err = testdb.CreateUser("TestReadUserByName_name", "TestReadUserByName_email", "TestReadUserByName_pass")
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
		err = testdb.CreateUser("ReadUserById_name", "ReadUserById_email", "ReadUserById_pass")
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

	for _, testdb := range testdatabases {
		usr, err := CreateTestUser(testdb)
		require.Nil(t, err)

		err = testdb.DeleteUser(usr.UserId)
		require.Nil(t, err)

		_, err = testdb.ReadUserById(usr.UserId)
		require.NotNil(t, err, "The user with ID %v should have errored out, but it did not", usr.UserId)

		err = testdb.DeleteUser(usr.UserId)
		require.Equal(t, err, ErrNothingToDelete, "Didn't catch try to delete deleted user")
	}
}

func TestReadUserDevice(t *testing.T) {

	for _, testdb := range testdatabases {
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

	for _, testdb := range testdatabases {
		user, err := CreateTestUser(testdb)

		_, _, err = testdb.Login(user.Name, testPassword)
		assert.Nil(t, err)

		_, _, err = testdb.Login(user.Email, testPassword)
		assert.Nil(t, err)

		_, _, err = testdb.Login("", testPassword)
		assert.Equal(t, err, ErrInvalidUsername, "Wrong type returned %v", err)

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
		assert.Equal(t, res, true, "Should want to upgrade a password with an old has type")

		assert.NotEqual(t, "", user.PasswordHashScheme, "The has scheme was not updated")
	}
}

func TestRevertUneditableFields(t *testing.T) {
	// The original value we're trying to change
	orig := User{1, "Name", "nick", "Email", "Password", "passsalt", "hash", true, 1, 1, 1}

	// the one we're trying to submit
	blank := User{0, "", "", "", "", "", "", false, 0, 0, 0}

	// nobody's version
	nobody := blank
	// root's version of blank:
	root := User{1, "", "", "", "", "", "", false, 0, 0, 0}
	// User's version of blank
	user := User{1, "Name", "", "", "", "", "", true, 1, 1, 1}
	// all the rest shouldn't be able to do anything
	device := orig
	family := orig
	enabled := orig
	anybody := orig

	tmpu := blank
	tmpu.RevertUneditableFields(orig, NOBODY)
	assert.Equal(t, tmpu, nobody, "Conversion as nobody didn't work got %v, expected %v", tmpu, root)

	tmpu = blank
	tmpu.RevertUneditableFields(orig, ROOT)
	assert.Equal(t, tmpu, root, "Conversion as root didn't work got %v, expected %v", tmpu, root)

	tmpu = blank
	tmpu.RevertUneditableFields(orig, USER)
	assert.Equal(t, tmpu, user, "Conversion as user didn't work got %v, expected %v", tmpu, root)

	tmpu = blank
	tmpu.RevertUneditableFields(orig, DEVICE)
	assert.Equal(t, tmpu, device, "Conversion as device didn't work got %v, expected %v", tmpu, root)

	tmpu = blank
	tmpu.RevertUneditableFields(orig, FAMILY)
	assert.Equal(t, tmpu, family, "Conversion as family didn't work got %v, expected %v", tmpu, root)

	tmpu = blank
	tmpu.RevertUneditableFields(orig, ENABLED)
	assert.Equal(t, tmpu, enabled, "Conversion as enabled didn't work got %v, expected %v", tmpu, root)

	tmpu = blank
	tmpu.RevertUneditableFields(orig, ANYBODY)
	assert.Equal(t, tmpu, anybody, "Conversion as anybody didn't work got %v, expected %v", tmpu, root)
}
