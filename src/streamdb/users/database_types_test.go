// Package users provides an API for managing user information.
package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetNewPassword(t *testing.T) {
	var j User
	var k User
	var l User

	j.SetNewPassword("monkey")
	k.SetNewPassword("password")
	l.PasswordSalt = "tmp"
	l.SetNewPassword("password")

	assert.NotEqual(t, j, k, "Setting password failed: %v vs %v", j, k)
	assert.NotEqual(t, k, l, "salting failed")

	j.SetNewPassword("password")
	assert.NotEqual(t, k, j, "second set failed, same hash")
}

func TestAdmin(t *testing.T) {
	var j User
	j.Admin = true
	var k User
	k.Admin = false

	assert.NotEqual(t, k.IsAdmin(), true, "false positive admin")
	assert.NotEqual(t, j.IsAdmin(), false, "false negative admin")
}

type ExpectedPemissions struct {
	in          string
	out         PermissionLevel
	errExpected bool
}

var permissionsTest = []ExpectedPemissions{
	{"nobody", NOBODY, false},
	{"root", ROOT, false},
	{"user", USER, false},
	{"device", DEVICE, false},
	{"family", FAMILY, false},
	{"enabled", ENABLED, false},
	{"anybody", ANYBODY, false},
	{"", ANYBODY, true}}

func TestStrToPermissionLevel(t *testing.T) {
	for _, test := range permissionsTest {
		pl, err := strToPermissionLevel(test.in)

		if test.errExpected && err == nil || !test.errExpected && err != nil {
			t.Errorf("Error failed for test %v", test.in)
		}

		if pl != test.out {
			t.Errorf("Wrong permission for %v, got %v expected %v", test.in, pl, test.out)
		}
	}
}

func TestDevicePermissions(t *testing.T) {
	var all Device
	all.IsAdmin = true
	all.Enabled = true
	all.CanWrite = true
	all.CanWriteAnywhere = true

	var none Device

	var onlyEnabled Device
	onlyEnabled.Enabled = true

	var disabledSuper Device
	disabledSuper.IsAdmin = true

	assert.False(t, onlyEnabled.IsAdmin, "improper elevation of privliges")
	assert.True(t, all.IsAdmin, "correct admin was denied")

	// CanModifyUser
	assert.False(t, none.CanActAsUser)
}

func TestGte(t *testing.T) {
	assert.True(t, NOBODY.Gte(ROOT))
	assert.True(t, ROOT.Gte(USER))
	assert.True(t, USER.Gte(DEVICE))
	assert.True(t, DEVICE.Gte(FAMILY))
	assert.True(t, FAMILY.Gte(ENABLED))
	assert.True(t, ENABLED.Gte(ANYBODY))
	assert.True(t, NOBODY.Gte(NOBODY))
}

func TestGeneralPermissions(t *testing.T) {
	var d1 Device
	d1.IsAdmin = true
	d1.Enabled = true
	assert.Equal(t, d1.GeneralPermissions(), ROOT)

	var d2 Device
	d2.Enabled = false
	assert.Equal(t, d2.GeneralPermissions(), ANYBODY)

	var d3 Device
	d3.IsAdmin = true
	d3.Enabled = false
	assert.Equal(t, d3.GeneralPermissions(), ANYBODY)

	var d4 Device
	d4.Enabled = true
	assert.Equal(t, d4.GeneralPermissions(), ENABLED)
}

func TestRelationToUser(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		u, dev, _, err := CreateUDS(testdb)
		require.Nil(t, err)

		assert.Equal(t, dev.RelationToUser(nil), ANYBODY, "nil user should be anybody")

		dev.Enabled = false
		assert.Equal(t, dev.RelationToUser(u), ANYBODY, "disabled should be anybody")
		dev.Enabled = true

		dev.IsAdmin = true
		assert.Equal(t, dev.RelationToUser(u), ROOT, "admin should have root")
		dev.IsAdmin = false

		dev.CanActAsUser = true
		assert.Equal(t, dev.RelationToUser(u), USER, "devices with same userid should be users")
		dev.CanActAsUser = false

		assert.Equal(t, dev.RelationToUser(u), DEVICE, "devices under a user should be a device")

		dev.UserId = -1
		assert.Equal(t, dev.RelationToUser(u), ANYBODY, "unrelated devices should be anybody")
	}
}

func TestRelationToDevice(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		u, dev, _, err := CreateUDS(testdb)
		require.Nil(t, err)

		d2, err := CreateTestDevice(testdb, u)
		require.Nil(t, err)

		assert.Equal(t, dev.RelationToDevice(nil), ANYBODY, "nil dev should be anybody")

		dev.Enabled = false
		assert.Equal(t, dev.RelationToDevice(d2), ANYBODY, "disabled devices should be anybody")
		dev.Enabled = true

		dev.IsAdmin = true
		assert.Equal(t, dev.RelationToDevice(d2), ROOT, "admin devices should have root")
		dev.IsAdmin = false

		dev.CanActAsUser = true
		assert.Equal(t, dev.RelationToDevice(d2), USER, "devices with same userid should be users")
		dev.CanActAsUser = false

		assert.Equal(t, dev.RelationToDevice(d2), FAMILY, "devices under a user should be a family")
		assert.Equal(t, dev.RelationToDevice(dev), DEVICE, "Devices should be device with themselves")

		dev.UserId = -1
		assert.Equal(t, dev.RelationToDevice(d2), ENABLED, "unrelated devices should be enabled")
	}
}

func TestRelationToStream(t *testing.T) {
	for i, testdb := range testdatabases {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		_, dev, stream, err := CreateUDS(testdb)
		require.Nil(t, err)

		assert.Equal(t, dev.RelationToStream(nil, dev), ANYBODY, "nil stream")
		assert.Equal(t, dev.RelationToStream(stream, nil), ANYBODY, "nil parent")

		dev.Enabled = false
		assert.Equal(t, dev.RelationToStream(stream, dev), ANYBODY, "disabled dev")
		dev.Enabled = true

		dev.IsAdmin = true
		assert.Equal(t, dev.RelationToStream(stream, dev), ROOT, "root dev")
		dev.IsAdmin = false

		dev.CanActAsUser = true
		assert.Equal(t, dev.RelationToStream(stream, dev), USER, "root dev")
		dev.CanActAsUser = false

		d2 := *dev
		d2.UserId = d2.UserId + 1
		d2.DeviceId = d2.DeviceId + 1
		assert.Equal(t, d2.RelationToStream(stream, dev), ENABLED,
			"different user devices got %v", dev.RelationToStream(stream, &d2))

		d2 = *dev
		d2.CanActAsUser = true
		d2.UserId += 1
		d2.DeviceId += 1
		assert.NotEqual(t, dev.RelationToStream(stream, &d2), USER, "d2 can be user, but isn't parent")
		assert.Equal(t, dev.RelationToStream(stream, dev), DEVICE, "owner should be dev")
	}
}
