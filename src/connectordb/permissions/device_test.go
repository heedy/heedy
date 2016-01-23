package permissions

import (
	pconfig "config/permissions"
	"connectordb/users"
	"testing"

	"github.com/stretchr/testify/require"
)

var testDevice2 = users.Device{
	Name:         "mydev2",
	UserID:       33,
	DeviceID:     1337,
	APIKey:       "notempty",
	Enabled:      true,
	IsVisible:    true,
	UserEditable: true,
}

func TestDeviceRead(t *testing.T) {
	u := &testUser
	d := &testDevice
	d2X := testDevice2
	d2 := &d2X

	// Other user does not have access to private devices
	m := ReadDeviceToMap(pconfig.Get(), u, d, d2)
	require.Nil(t, m)

	// Other user DOES have read access to public devices
	d2.Public = true
	m = ReadDeviceToMap(pconfig.Get(), u, d, d2)
	require.NotNil(t, m)
	v, ok := m["name"]
	require.True(t, ok)
	require.Equal(t, "mydev2", v.(string))
	_, ok = m["apikey"]
	require.False(t, ok)

	// Now check what happens if our owner is in fact us
	d2.Public = false
	d2.UserID = 32
	m = ReadDeviceToMap(pconfig.Get(), u, d, d2)
	require.NotNil(t, m)
	v, ok = m["apikey"]
	require.True(t, ok)
	require.Equal(t, "notempty", v.(string))
}

func TestDeviceWrite(t *testing.T) {
	u := &testUser
	d := &testDevice
	d2X := testDevice2
	d2 := &d2X

	// No access to the device
	require.Error(t, UpdateDeviceFromMap(pconfig.Get(), u, d, d2, map[string]interface{}{"nickname": "hi"}))

	// No WRITE access to the device
	d2.Public = true
	require.Error(t, UpdateDeviceFromMap(pconfig.Get(), u, d, d2, map[string]interface{}{"nickname": "hi"}))

	// We own the device
	d2.UserID = 32
	require.NoError(t, UpdateDeviceFromMap(pconfig.Get(), u, d, d2, map[string]interface{}{"nickname": "hi"}))
	require.Equal(t, "hi", d2.Nickname)

	// No changing permissions of a user device
	d2.Name = "user"
	require.Error(t, UpdateDeviceFromMap(pconfig.Get(), u, d, d2, map[string]interface{}{"can_read_user": false}))

	// But it's fine if it isnt a user device
	d2.Name = "mydevice2"
	require.NoError(t, UpdateDeviceFromMap(pconfig.Get(), u, d, d2, map[string]interface{}{"can_read_user": true}))
	require.True(t, d2.CanReadUser)

	// We can't update the name of a device
	require.Error(t, UpdateDeviceFromMap(pconfig.Get(), u, d, d2, map[string]interface{}{"name": "hi"}))
}
