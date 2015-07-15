package streamdb

import (
	"connectordb/config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseDeviceCrud(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()
	//go db.RunWriter()

	_, err = db.ReadAllDevices("notauser")
	require.Error(t, err)

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	devs, err := db.ReadAllDevices("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, 1, len(devs)) //the user device

	_, err = db.ReadDevice("streamdb_test/testdevice")
	require.Error(t, err)

	require.NoError(t, db.CreateDevice("streamdb_test/testdevice"))

	devs, err = db.ReadAllDevices("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, 2, len(devs)) //should have our new device

	dev, err := db.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.Equal(t, "testdevice", dev.Name)

	key, err := db.ChangeDeviceAPIKey("streamdb_test/testdevice")
	require.NoError(t, err)
	require.NotEqual(t, key, dev.ApiKey)
	dev, err = db.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.Equal(t, key, dev.ApiKey)

	require.NoError(t, db.DeleteDevice("streamdb_test/testdevice"))

	_, err = db.ReadDevice("streamdb_test/testdevice")
	require.Error(t, err)

	require.NoError(t, db.CreateDevice("streamdb_test/testdevice2"))
	require.NoError(t, db.CreateDevice("streamdb_test/testdevice1"))
	require.NoError(t, db.CreateDevice("streamdb_test/testdevice3"))

	require.NoError(t, db.SetAdmin("streamdb_test/testdevice3", true))
	require.Error(t, db.SetAdmin("streamdb_test/testdevice4", true))

	dev, err = db.ReadDevice("streamdb_test/testdevice1")
	//Clear the cache
	db.Reload()
	dev.Name = "hiah"
	require.NoError(t, db.UpdateDevice(dev))

}
