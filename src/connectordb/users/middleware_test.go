/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"reflect"
	"testing"
)

type MiddlewareTestcase struct {
	// The database that will provide the baseline results
	Base UserDatabase
	// The database we're testing
	Test UserDatabase
	// The expected number of database calls
	NumCalls uint64
}

var (
	ErrBackend   ErrorUserdb
	KnownBackend KnownUserdb
	ErrUserdb    UserDatabase = &ErrBackend
)

func AssertEqMiddlewareTest(t *testing.T, testResult, expectedResult interface{}, testname string, testindex int) {
	if reflect.DeepEqual(testResult, expectedResult) {
		return
	}

	t.Errorf("Error in %s test# %d | expected: %v got: %v", testname, testindex, expectedResult, testResult)
}

func GetCommonTestcases() []MiddlewareTestcase {
	cacheTest1, _ := NewCacheMiddleware(&ErrBackend, 100, 100, 100)
	cacheTest2, _ := NewCacheMiddleware(&KnownBackend, 100, 100, 100)
	return []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, cacheTest1, 1},
		{&KnownBackend, cacheTest2, 1}}
}

func TestMiddlewareCreateDevice(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.CreateDevice("", 0, 0)
		baseError := testcase.Base.CreateDevice("", 0, 0)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareCreateDevice Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareCreateDevice #Calls", index)
	}
}

func TestMiddlewareCreateStream(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.CreateStream("", "", 0, 0)
		baseError := testcase.Base.CreateStream("", "", 0, 0)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareCreateStream Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareCreateStreams #Calls", index)
	}
}

func TestMiddlewareCreateUser(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.CreateUser("", "", "", "", 0)
		baseError := testcase.Base.CreateUser("", "", "", "", 0)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareCreateUser Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareCreateUser #Calls", index)
	}
}

func TestMiddlewareDeleteUser(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.DeleteUser(1)
		baseError := testcase.Base.DeleteUser(1)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareDeleteUser Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareDeleteUser #Calls", index)
	}
}

func TestMiddlewareDeleteDevice(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.DeleteDevice(1)
		baseError := testcase.Base.DeleteDevice(1)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareDeleteDevice Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareDeleteDevice #Calls", index)
	}
}

func TestMiddlewareDeleteStream(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.DeleteStream(1)
		baseError := testcase.Base.DeleteStream(1)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareDeleteStream Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareDeleteStream #Calls", index)
	}
}

func TestMiddlewareUpdateUser(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.UpdateUser(&User{})
		baseError := testcase.Base.UpdateUser(&User{})

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareUpdateUser Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareUpdateUser #Calls", index)
	}
}

func TestMiddlewareUpdateDevice(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.UpdateDevice(&Device{})
		baseError := testcase.Base.UpdateDevice(&Device{})

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareUpdateDevice Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareUpdateDevice #Calls", index)
	}
}

func TestMiddlewareUpdateStream(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.UpdateStream(&Stream{})
		baseError := testcase.Base.UpdateStream(&Stream{})

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareUpdateStream Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareUpdateStream #Calls", index)
	}
}

func TestMiddlewareReadAllUsers(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadAllUsers()
		baseResult, baseError := testcase.Base.ReadAllUsers()

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareReadUsers Errors", index)
		AssertEqMiddlewareTest(t, testResult, baseResult, "TestMiddlewareReadUsers Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareUpdateStream #Calls", index)
	}
}

func TestMiddlewareReadDeviceByAPIKey(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadDeviceByAPIKey("")
		baseResult, baseError := testcase.Base.ReadDeviceByAPIKey("")

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadDeviceByAPIKey"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadDeviceByID(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadDeviceByID(0)
		baseResult, baseError := testcase.Base.ReadDeviceByID(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadDeviceByID"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadUserOperatingDevice(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadUserOperatingDevice(&User{})
		baseResult, baseError := testcase.Base.ReadUserOperatingDevice(&User{})

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadUserOperatingDevice"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadReadDeviceForUserByName(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadDeviceForUserByName(0, "")
		baseResult, baseError := testcase.Base.ReadDeviceForUserByName(0, "")

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadDeviceForUserByName"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, testResult, baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadDeviceForUserID(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadDevicesForUserID(0)
		baseResult, baseError := testcase.Base.ReadDevicesForUserID(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadDeviceForUserID"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, testResult, baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadStreamByDeviceIDAndName(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadStreamByDeviceIDAndName(0, "foo")
		baseResult, baseError := testcase.Base.ReadStreamByDeviceIDAndName(0, "foo")

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadStreamByDeviceIDAndName"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadStreamByID(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadStreamByID(0)
		baseResult, baseError := testcase.Base.ReadStreamByID(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadStreamByID"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadStreamByDevice(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadStreamsByDevice(0)
		baseResult, baseError := testcase.Base.ReadStreamsByDevice(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadStreamByDevice"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, testResult, baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadUserById(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadUserById(0)
		baseResult, baseError := testcase.Base.ReadUserById(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadUserById"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadUserByName(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadUserByName("")
		baseResult, baseError := testcase.Base.ReadUserByName("")

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadUserByName"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareLogin(t *testing.T) {
	var testcases = GetCommonTestcases()

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testUser, testDevice, testError := testCounter.Login("", "")
		baseUser, baseDevice, baseError := testcase.Base.Login("", "")

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareLogin"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testUser, &baseUser, prefix+" User", index)
		AssertEqMiddlewareTest(t, &testDevice, &baseDevice, prefix+" Device", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}
