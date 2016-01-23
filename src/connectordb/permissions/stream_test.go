package permissions

import (
	pconfig "config/permissions"
	"connectordb/users"
	"testing"

	"github.com/stretchr/testify/require"
)

var testStream = users.Stream{
	Name:     "mystream",
	DeviceID: 1337,
	Schema:   `{"type":"number"}`,
	StreamID: 9001,
}

func TestStreamRead(t *testing.T) {
	u := &testUser
	d := &testDevice
	d2X := testDevice2
	d2 := &d2X
	sX := testStream
	s := &sX

	// Other user does not have access to private devices
	m := ReadStreamToMap(pconfig.Get(), u, d, d2, s)
	require.Nil(t, m)

	// Other user DOES have read access to public devices
	d2.Public = true
	m = ReadStreamToMap(pconfig.Get(), u, d, d2, s)
	require.NotNil(t, m)
	v, ok := m["name"]
	require.True(t, ok)
	require.Equal(t, "mystream", v.(string))
}

func TestStreamWrite(t *testing.T) {
	u := &testUser
	d := &testDevice
	d2X := testDevice2
	d2 := &d2X
	sX := testStream
	s := &sX

	// No access to the stream
	require.Error(t, UpdateStreamFromMap(pconfig.Get(), u, d, d2, s, map[string]interface{}{"nickname": "hi"}))

	// No WRITE access to the stream
	d2.Public = true
	require.Error(t, UpdateStreamFromMap(pconfig.Get(), u, d, d2, s, map[string]interface{}{"nickname": "hi"}))
	// We own the device
	d2.UserID = 32
	require.NoError(t, UpdateStreamFromMap(pconfig.Get(), u, d, d2, s, map[string]interface{}{"nickname": "hi"}))
	require.Equal(t, "hi", s.Nickname)

	// Schema update not allowed
	require.Error(t, UpdateStreamFromMap(pconfig.Get(), u, d, d2, s, map[string]interface{}{"schema": `{"type":"string"}`}))
	require.Error(t, UpdateStreamFromMap(pconfig.Get(), u, d, d2, s, map[string]interface{}{"name": `hi`}))
}
