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

func TestMiddlewareCreateDevice(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.CreateDevice("", 0)
		baseError := testcase.Base.CreateDevice("", 0)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareCreateDevice Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareCreateDevice #Calls", index)
	}
}

func TestMiddlewareCreateStream(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.CreateStream("", "", 0)
		baseError := testcase.Base.CreateStream("", "", 0)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareCreateStream Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareCreateStreams #Calls", index)
	}
}

func TestMiddlewareCreateUser(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.CreateUser("", "", "")
		baseError := testcase.Base.CreateUser("", "", "")

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareCreateUser Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareCreateUser #Calls", index)
	}
}

func TestMiddlewareDeleteUser(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.UpdateUser(nil)
		baseError := testcase.Base.UpdateUser(nil)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareUpdateUser Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareUpdateUser #Calls", index)
	}
}

func TestMiddlewareUpdateDevice(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.UpdateDevice(nil)
		baseError := testcase.Base.UpdateDevice(nil)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareUpdateDevice Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareUpdateDevice #Calls", index)
	}
}

func TestMiddlewareUpdateStream(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testError := testCounter.UpdateStream(nil)
		baseError := testcase.Base.UpdateStream(nil)

		numCalls := testCounter.GetNumberOfCalls()

		AssertEqMiddlewareTest(t, testError, baseError, "TestMiddlewareUpdateStream Errors", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, "TestMiddlewareUpdateStream #Calls", index)
	}
}

func TestMiddlewareReadAllUsers(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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

func TestMiddlewareReadDeviceByApiKey(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadDeviceByApiKey("")
		baseResult, baseError := testcase.Base.ReadDeviceByApiKey("")

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadDeviceByApiKey"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadDeviceById(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadDeviceById(0)
		baseResult, baseError := testcase.Base.ReadDeviceById(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadDeviceById"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadUserOperatingDevice(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadUserOperatingDevice(nil)
		baseResult, baseError := testcase.Base.ReadUserOperatingDevice(nil)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadUserOperatingDevice"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadReadDeviceForUserByName(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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

func TestMiddlewareReadDeviceForUserId(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadDevicesForUserId(0)
		baseResult, baseError := testcase.Base.ReadDevicesForUserId(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadDeviceForUserId"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, testResult, baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadStreamByDeviceIdAndName(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadStreamByDeviceIdAndName(0, "foo")
		baseResult, baseError := testcase.Base.ReadStreamByDeviceIdAndName(0, "foo")

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadStreamByDeviceIdAndName"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadStreamById(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

	for index, testcase := range testcases {
		testCounter := AccountingMiddleware{testcase.Test, 0}
		testResult, testError := testCounter.ReadStreamById(0)
		baseResult, baseError := testcase.Base.ReadStreamById(0)

		numCalls := testCounter.GetNumberOfCalls()

		prefix := "TestMiddlewareReadStreamById"
		AssertEqMiddlewareTest(t, testError, baseError, prefix+" Errors", index)
		AssertEqMiddlewareTest(t, &testResult, &baseResult, prefix+" Result", index)
		AssertEqMiddlewareTest(t, numCalls, testcase.NumCalls, prefix+" #Calls", index)
	}
}

func TestMiddlewareReadStreamByDevice(t *testing.T) {
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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
	var testcases = []MiddlewareTestcase{
		{&ErrBackend, &IdentityMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &IdentityMiddleware{&KnownBackend}, 1},
		{&ErrBackend, &CacheMiddleware{&ErrBackend}, 1},
		{&KnownBackend, &CacheMiddleware{&KnownBackend}, 1}}

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
