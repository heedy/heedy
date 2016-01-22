package permissions

import (
	"config"
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
	require.NotNil(t, ReadUserToMap(&config.TestConfiguration, u, d, mu))

	d.CanReadExternal = false
	require.Nil(t, ReadUserToMap(&config.TestConfiguration, u, d, mu))
	mu.Public = false
	require.Nil(t, ReadUserToMap(&config.TestConfiguration, u, d, mu))
	require.NotNil(t, ReadUserToMap(&config.TestConfiguration, u, d, u))

	d.CanReadUser = false
	require.Nil(t, ReadUserToMap(&config.TestConfiguration, u, d, mu))
}

func TestDeviceWritePermissions(t *testing.T) {
	u := &testUser
	muX := testUser2
	mu := &muX
	dX := testDevice
	d := &dX
	mu.Permissions = "admin"
	require.NoError(t, UpdateUserFromMap(&config.TestConfiguration, mu, d, mu, map[string]interface{}{"nickname": "hi"}))

	d.CanWriteUser = false
	require.Error(t, UpdateUserFromMap(&config.TestConfiguration, mu, d, mu, map[string]interface{}{"nickname": "hi"}))

	d.CanWriteExternal = false
	require.Error(t, UpdateUserFromMap(&config.TestConfiguration, mu, d, u, map[string]interface{}{"nickname": "hi"}))
}
