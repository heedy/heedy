package connectordb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevice(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	num, err := db.CountDevices()
	require.NoError(t, err)
	require.Equal(t, int64(0), num)

	u, err := db.ReadDevice("tst/tst")
	require.Nil(t, u)
	require.Error(t, err)

	require.NoError(t, db.CreateUser("myuser", "email@email", "test", "user", true))

	require.Error(t, db.CreateDevice("nouser/mydevice"))
	require.NoError(t, db.CreateDevice("myuser/mydevice"))
	require.Error(t, db.CreateDevice("myuser/mydevice"))

	u, err = db.ReadDevice("myuser/mydevice")
	require.NoError(t, err)
	require.Equal(t, "mydevice", u.Name)
	require.Equal(t, false, u.Public)
	require.Equal(t, "", u.Role)

	require.NoError(t, db.DeleteDevice("myuser/mydevice"))

	_, err = db.ReadUser("myuser/mydevice")
	require.Error(t, err)

	require.Error(t, db.DeleteDevice("myuser/mydevice"))

	require.NoError(t, db.CreateDevice("myuser/mydevice"))
	u, err = db.ReadDevice("myuser/mydevice")
	require.NoError(t, err)
	require.NoError(t, db.DeleteUser("myuser"))
	_, err = db.ReadDeviceByID(u.DeviceID)
	require.Error(t, err)
}
