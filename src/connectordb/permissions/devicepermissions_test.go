package permissions

import (
	pconfig "config/permissions"
	"testing"

	"github.com/stretchr/testify/require"
)

// Some variables are defined in user_test.go

func TestDeviceReadPermissions(t *testing.T) {
	u := &testUser
	muX := testUser2
	mu := &muX
	dX := testDevice
	d := &dX
	require.NotNil(t, ReadUserToMap(pconfig.Get(), u, d, mu))

	d.CanReadExternal = false
	require.Nil(t, ReadUserToMap(pconfig.Get(), u, d, mu))
	mu.Public = false
	require.Nil(t, ReadUserToMap(pconfig.Get(), u, d, mu))
	require.NotNil(t, ReadUserToMap(pconfig.Get(), u, d, u))

	d.CanReadUser = false
	require.Nil(t, ReadUserToMap(pconfig.Get(), u, d, mu))
}

func TestDeviceWritePermissions(t *testing.T) {
	u := &testUser
	muX := testUser2
	mu := &muX
	dX := testDevice
	d := &dX
	mu.Role = "admin"
	require.NoError(t, UpdateUserFromMap(pconfig.Get(), mu, d, mu, map[string]interface{}{"nickname": "hi"}))

	d.CanWriteUser = false
	require.Error(t, UpdateUserFromMap(pconfig.Get(), mu, d, mu, map[string]interface{}{"nickname": "hi"}))

	d.CanWriteExternal = false
	require.Error(t, UpdateUserFromMap(pconfig.Get(), mu, d, u, map[string]interface{}{"nickname": "hi"}))
}
