package users

import (
	"testing"
	"reflect"
)


func TestCreateUser(t *testing.T) {
	CleanTestDB()

	err := testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email", "TestCreateUser_pass")
	if err != nil {
		t.Errorf("Cannot create user %v", err)
		return
	}

	err = testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email2", "TestCreateUser_pass2")
	if err == nil {
		t.Errorf("Wrong error returned %v", err)
		return
	}

	err = testdb.CreateUser("TestCreateUser_name2", "TestCreateUser_email", "TestCreateUser_pass2")
	if err == nil {
		t.Errorf("Wrong err returned %v", err)
		return
	}
}

func BenchmarkCreateUser(b *testing.B) {
    for i := 0; i < b.N; i++ {
        testdb.CreateUser(GetNextName(), GetNextEmail(), "TestCreateUser_pass2")
    }
}


func TestReadAllUsers(t *testing.T) {
	CleanTestDB()

	for i := 0; i < 5; i++{
		_, err := CreateTestUser()
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}

	users, err := testdb.ReadAllUsers()

	if err != nil {
		t.Errorf("Exception while reading all users %v", err)
	}

	if users == nil {
		t.Errorf("users is nil")
		return
	}

	err = testdb.CreateUser("TestReadAllUsers", "TestReadAllUsers_email", "TestReadAllUsers_pass")
	if err != nil {
		t.Errorf("Could not complete test due to: %v", err)
		return
	}

	users2, err := testdb.ReadAllUsers()

	if err != nil {
		t.Errorf("Exception while reading all users %v", err)
	}

	if users2 == nil {
		t.Errorf("users is nil")
		return
	}

	if len(users2)-len(users) != 1 {
		t.Errorf("not taking into account changes")
	}
}

func TestReadUserByEmail(t *testing.T) {
	// test failures on non existance
	usr, err := testdb.ReadUserByEmail("doesnotexist   because spaces")

	if err == nil {
		t.Errorf("no error returned, expected non nil on failing case")
	}

	// setup for reading
	err = testdb.CreateUser("TestReadUserByEmail_name", "TestReadUserByEmail_email", "TestReadUserByEmail_pass")
	if err != nil {
		t.Errorf("Could not create user for test reading... %v", err)
		return
	}

	usr, err = testdb.ReadUserByEmail("TestReadUserByEmail_email")

	if usr == nil {
		t.Errorf("did not get a user by email")
	}

	if err != nil {
		t.Errorf("got an error when trying to get a user that should exist %v", err)
	}
}

func TestReadUserByName(t *testing.T) {
	// test failures on non existance
	usr, err := testdb.ReadUserByName("")

	if err == nil {
		t.Errorf("no error returned, expected non nil on failing case")
	}

	// setup for reading
	err = testdb.CreateUser("TestReadUserByName_name", "TestReadUserByName_email", "TestReadUserByName_pass")
	if err != nil {
		t.Errorf("Could not create user for test reading... %v", err)
		return
	}

	usr, err = testdb.ReadUserByName("TestReadUserByName_name")

	if usr == nil {
		t.Errorf("did not get a user by name")
	}

	if err != nil {
		t.Errorf("got an error when trying to get a user that should exist %v", err)
	}
}

func TestReadUserById(t *testing.T) {
	// test failures on non existance
	usr, err := testdb.ReadUserById(-1)

	if err == nil {
		t.Errorf("no error returned, expected non nil on failing case")
	}

	// setup for reading
	err = testdb.CreateUser("ReadUserById_name", "ReadUserById_email", "ReadUserById_pass")
	if err != nil {
		t.Errorf("Could not create user for test reading... %v", err)
		return
	}

	usr, err = testdb.ReadUserByName("ReadUserById_name")

	if usr == nil {
		t.Errorf("did not get a user by id")
	}

	if err != nil {
		t.Errorf("got an error when trying to get a user that should exist %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	usr, err := CreateTestUser()

	if err != nil {
		t.Errorf("got an error when trying to create a user %v", err.Error())
		return
	}

	err = testdb.UpdateUser(nil)
	if err != ERR_INVALID_PTR {
		t.Errorf("Didn't catch nil")
	}

	usr.Name = "Hello"
	usr.Email = "hello@example.com"
	usr.Admin = true
	usr.UploadLimit_Items = 1
	usr.ProcessingLimit_S = 1
	usr.StorageLimit_Gb = 1

	err = testdb.UpdateUser(usr)

	if err != nil {
		t.Errorf("Could not update user %v", err)
	}

	usr2, err := testdb.ReadUserByName(usr.Name)

	if !reflect.DeepEqual(usr, usr2) {
		t.Errorf("The original and updated objects don't match orig: %v updated %v", usr, usr2)
	}
}

func TestDeleteUser(t *testing.T) {
	usr, err := CreateTestUser()

	if nil != err {
		t.Errorf("Cannot create user to test delete")
		return
	}

	err = testdb.DeleteUser(usr.UserId)

	if nil != err {
		t.Errorf("Error when attempted delete %v", err)
		return
	}

	_, err = testdb.ReadUserById(usr.UserId)

	if err == nil {
		t.Errorf("The user with ID %v should have errored out, but it did not", usr.UserId)
		return
	}
}

func TestReadStreamOwner(t *testing.T) {
	user, err := CreateTestUser()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	device, err := CreateTestDevice(user)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	stream, err := CreateTestStream(device)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	owner, err := testdb.ReadStreamOwner(stream.StreamId)

	if err != nil {
		t.Errorf("Could not read stream owner %v", err)
	}

	if owner.UserId != user.UserId {
		t.Errorf("Wrong stream owner got %v, expected %v", owner, user)
	}
}

func TestReadUserDevice(t *testing.T) {
	user, err := CreateTestUser()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	dev, err := testdb.ReadUserOperatingDevice(user)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if dev.UserId != user.UserId {
		t.Errorf("Incorrect device returned.")
	}
}

func TestLogin(t *testing.T) {

	user, err := CreateTestUser()

	_, _, err = testdb.Login(user.Name, TEST_PASSWORD)
	if err != nil {
		t.Errorf(err.Error())
	}

	_, _, err = testdb.Login(user.Email, TEST_PASSWORD)
	if err != nil {
		t.Errorf(err.Error())
	}


	_, _, err = testdb.Login("", TEST_PASSWORD)
	if err != InvalidUsernameError {
		t.Errorf("Wrong type returned %v", err)
	}

	_, _, err = testdb.Login(user.Name, "")
	if err == nil {
		t.Errorf("Accepted blank password")
	}

}

func TestUpgradePassword(t *testing.T) {
	user, err := CreateTestUser()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	res := user.UpgradePassword(TEST_PASSWORD)
	if res != false {
		t.Errorf("Should not need to upgrade a password with the same salt")
	}

	user.PasswordHashScheme = ""

	res = user.UpgradePassword(TEST_PASSWORD)
	if res != true {
		t.Errorf("Should want to upgrade a password with an old has type")
	}

	if user.PasswordHashScheme == "" {
		t.Errorf("The has scheme was not updated")
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
	if tmpu != root {
		t.Errorf("Conversion as root didn't work got %v, expected %v", tmpu, root)
	}

	tmpu = orig
	tmpu.RevertUneditableFields(blank, USER)
	if tmpu != user {
		t.Errorf("Conversion as user didn't work got %v, expected %v", tmpu, root)
	}

}
/*
// User is the storage type for rows of the database.
type User struct {
        UserId    int64  `modifiable:"nobody"` // The primary key
        Name  string `modifiable:"root"`   // The public username of the user
        Email string `modifiable:"user"`   // The user's email address

        Password           string `modifiable:"user"` // A hash of the user's password
        PasswordSalt       string `modifiable:"user"` // The password salt to be attached to the end of the password
        PasswordHashScheme string `modifiable:"user"` // A string representing the hashing scheme used

        Admin        bool   `modifiable:"root"` // True/False if this is an administrator

        UploadLimit_Items int `modifiable:"root"` // upload limit in items/day
        ProcessingLimit_S int `modifiable:"root"` // processing limit in seconds/day
        StorageLimit_Gb   int `modifiable:"root"` // storage limit in GB
}

func (d *User) RevertUneditableFields(originalValue User, p PermissionLevel) {
        revertUneditableFields(d, originalValue, p)
}
*/
