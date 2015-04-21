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

func TestValidateUser(t *testing.T) {
	name, email, pass := "TestValidateUser_name", "TestValidateUser_email", "TestValidateUser_pass"

	err := testdb.CreateUser(name, email, pass)
	if err != nil {
		t.Errorf("Cannot create user %v", err)
		return
	}

	validated, _ := testdb.ValidateUser(name, pass)
	if !validated {
		t.Errorf("could not validate a user with username and pass")
	}

	validated, _ = testdb.ValidateUser(email, pass)
	if ! validated {
		t.Errorf("could not validate a user with email and pass")
	}

	validated, _ = testdb.ValidateUser(email, email)
	if validated {
		t.Errorf("Validated an incorrect user")
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
