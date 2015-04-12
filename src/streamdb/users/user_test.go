package users

import (
	"testing"
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
/**
func TestReadUserByEmail(t *testing.T) {
	// test failures on non existance
	usr, err := testdb.ReadUserByEmail("doesnotexist   because spaces")

	if usr != nil {
		t.Errorf("Selected user that does not exist by email")
	}

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

	if usr != nil {
		t.Errorf("Selected user that does not exist by name")
	}

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
	if !validated {
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

	if usr != nil {
		t.Errorf("Selected user that does not exist by name")
	}

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
/**
func TestUpdateUser(t *testing.T) {
	// setup for reading
	id, err := testdb.CreateUser("TestUpdateUser_name", "TestUpdateUser_email", "TestUpdateUser_pass")
	if err != nil {
		t.Errorf("Could not create user for test reading... %v", err)
		return
	}

	usr, err := testdb.ReadUserById(id)

	if usr == nil {
		t.Errorf("did not get a user by id")
		return
	}

	if err != nil {
		t.Errorf("got an error when trying to get a user that should exist %v", err)
		return
	}

	err = testdb.UpdateUser(nil)
	if err != ERR_INVALID_PTR {
		t.Errorf("Didn't catch nil")
	}

	usr.Name = "Hello"
	usr.Email = "hello@example.com"
	usr.Admin = true
	usr.Phone = "(303) 303-0000" //Non-legal phone number, don't worry
	usr.UploadLimit_Items = 1
	usr.ProcessingLimit_S = 1
	usr.StorageLimit_Gb = 1

	err = testdb.UpdateUser(usr)

	if err != nil {
		t.Errorf("Could not update user %v", err)
	}

	usr2, err := testdb.ReadUserById(id)

	usr.ModifyTime = usr2.ModifyTime // have to do this because we update it each time!
	if err != nil {
		t.Errorf("got an error when trying to get a user that should exist %v", err)
		return
	}

	if !reflect.DeepEqual(usr, usr2) {
		t.Errorf("The original and updated objects don't match orig: %v updated %v", usr, usr2)
	}
}

func TestConstructUserFrom(t *testing.T) {
	teste := errors.New("blah")
	_, err := constructUserFromRow(nil, teste)

	if err != teste {
		t.Errorf("Construct user did not allow error passthrough.")
	}

	// Different Method

	_, err = constructUsersFromRows(nil)

	if err != ERR_INVALID_PTR {
		t.Errorf("allowed an illegal pointer through")
	}
}

func TestDeleteUser(t *testing.T) {
	id, err := testdb.CreateUser("a", "b", "c")

	if nil != err {
		t.Errorf("Cannot create user to test delete")
		return
	}

	err = testdb.DeleteUser(id)

	if nil != err {
		t.Errorf("Error when attempted delete %v", err)
		return
	}

	user, err := testdb.ReadUserById(id)

	if err == nil {
		t.Errorf("The user with ID %v should have errored out, but it did not", id)
		return
	}
	if user != nil {
		t.Errorf("Expected nil, but we got back a user meaning the delete failed %v", user)
	}
}

func TestReadStreamOwner(t *testing.T) {
	id, err := testdb.CreateUser("TestReadStreamOwner_name", "TestReadStreamOwner_email", "TestReadStreamOwner_pass")
	if err != nil {
		t.Errorf("Could not create user for test owners... %v", err)
		return
	}

	user, err := testdb.ReadUserById(id)
	if err != nil || user == nil {
		t.Errorf("The user with ID %v does not exist or err %v", id, err)
		return
	}

	devid, err := testdb.CreateDevice("Devname", user)
	if err != nil {
		t.Errorf("Could not create device for test owners... %v", err)
		return
	}

	device, err := testdb.ReadDeviceById(devid)
	if err != nil || device == nil {
		t.Errorf("The device with ID %v does not exist or err %v", id, err)
		return
	}

	streamid, _ := testdb.CreateStream("TestReadStreamOwner", "", device)

	owner, err := testdb.ReadStreamOwner(streamid)

	if err != nil {
		t.Errorf("Could not read stream owner %v", err)
	}

	if owner.Id != user.Id {
		t.Errorf("Wrong stream owner got %v, expected %v", owner, user)
	}
}
**/
