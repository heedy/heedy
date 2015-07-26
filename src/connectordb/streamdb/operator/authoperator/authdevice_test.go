package authoperator

import (
	"testing"

	"connectordb/config"

	"github.com/stretchr/testify/require"
)

func TestAuthDeviceUserCrud(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()
	//go db.RunWriter()

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))
	require.NoError(t, db.CreateUser("otheruser", "root@localhost2", "mypass"))
	require.NoError(t, db.CreateDevice("otheruser/testdevice"))
	testdevice, err := db.ReadDevice("otheruser/testdevice")
	require.NoError(t, err)

	o, err := db.GetOperator("streamdb_test")
	require.NoError(t, err)

	devs, err := o.ReadAllDevices("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, 1, len(devs)) //the user device

	dev, err := o.Device()
	require.NoError(t, err)

	o2, err := db.DeviceOperator(dev.DeviceId)
	require.NoError(t, err)
	require.Equal(t, "streamdb_test/user", o2.Name())

	devs, err = o.ReadAllDevicesByUserID(dev.UserId)
	require.NoError(t, err)
	require.Equal(t, 1, len(devs)) //the user device

	//This user should not be able to CRUD devices of another user
	devs, err = o.ReadAllDevices("otheruser")
	require.Error(t, err)
	devs, err = o.ReadAllDevicesByUserID(5436)
	require.Error(t, err)

	_, err = o.ReadDevice("otheruser/testdevice")
	require.Error(t, err)
	require.Error(t, o.DeleteDevice("otheruser/testdevice"))
	require.Error(t, o.CreateDevice("otheruser/testdevice2"))
	_, err = db.ReadDevice("otheruser/testdevice2")
	require.Error(t, err)

	dev, err = db.ReadDevice("otheruser/testdevice")
	require.NoError(t, err)
	_, err = o.ReadDeviceByID(dev.DeviceId)
	require.Error(t, err)
	_, err = o.ReadDeviceByUserID(dev.UserId, "testdevice")
	require.Error(t, err)
	require.Error(t, o.DeleteDeviceByID(dev.DeviceId))
	require.Error(t, o.CreateDeviceByUserID(dev.UserId, "testdevice2"))

	testdevice.Nickname = "test"
	require.Error(t, o.UpdateDevice(testdevice))

	require.Error(t, o.SetAdmin("otheruser/testdevice", true))
	_, err = o.ChangeDeviceAPIKey("otheruser/testdevice")
	require.Error(t, err)

	//This user should be able to crud its own devices
	require.NoError(t, o.CreateDevice("streamdb_test/testdevice"))
	dev, err = o.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.Equal(t, "testdevice", dev.Name)
	dev, err = o.ReadDeviceByID(dev.DeviceId)
	require.NoError(t, err)
	require.Equal(t, "testdevice", dev.Name)
	dev, err = o.ReadDeviceByUserID(dev.UserId, "testdevice")
	require.NoError(t, err)
	require.Equal(t, "testdevice", dev.Name)

	_, err = db.LoginOperator("streamdb_test/testdevice", dev.ApiKey)
	require.NoError(t, err)
	oldkey := dev.ApiKey

	key, err := o.ChangeDeviceAPIKey("streamdb_test/testdevice")
	require.NoError(t, err)
	require.NotEqual(t, key, dev.ApiKey)
	dev, err = o.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.Equal(t, key, dev.ApiKey)

	_, err = db.LoginOperator("streamdb_test/testdevice", oldkey)
	require.Error(t, err)

	require.NoError(t, o.DeleteDevice("streamdb_test/testdevice"))

	usr, err := o.User()
	require.NoError(t, o.CreateDeviceByUserID(usr.UserId, "testdevice"))
	dev, err = o.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.NoError(t, o.DeleteDeviceByID(dev.DeviceId))
}

func TestAuthDeviceDeviceCrud(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()
	//go db.RunWriter()

	require.NoError(t, db.CreateUser("tstusr", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tstusr/testdevice"))
	require.NoError(t, db.CreateDevice("tstusr/test"))

	o, err := db.GetOperator("tstusr/test")
	require.NoError(t, err)

	//This device should not be able to CRUD other devices
	_, err = o.ReadDevice("tstusr/testdevice")
	require.Error(t, err)
	require.Error(t, o.DeleteDevice("tstusr/testdevice"))
	require.Error(t, o.CreateDevice("tstusr/testdevice2"))
	_, err = db.ReadDevice("tstusr/testdevice2")
	require.Error(t, err)

	testdevice, err := db.ReadDevice("tstusr/testdevice")
	require.NoError(t, err)
	testdevice.Nickname = "test"
	require.Error(t, o.UpdateDevice(testdevice))

	require.Error(t, o.SetAdmin("tstusr/testdevice", true))
	_, err = o.ChangeDeviceAPIKey("tstusr/testdevice")
	require.Error(t, err)

	u, err := o.User()
	require.NoError(t, err)
	u.Email = "changedemail@lol"
	require.Error(t, o.UpdateUser(u))

	//This device should be able to modify itself
	dev, err := o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	require.Equal(t, "test", dev.Name)

	//Shouldn't have those permissions
	require.Error(t, o.SetAdmin("tstusr/test", true))

	//Lastly, shouldn't be able to self-userify
	dev.CanActAsUser = true
	require.Error(t, o.UpdateDevice(dev))

	dev, err = o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	//But changing nickname is fine
	dev.Nickname = "testnick"
	require.NoError(t, o.UpdateDevice(dev))
	dev, err = o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	require.Equal(t, "testnick", dev.Nickname)

	key, err := o.ChangeDeviceAPIKey("tstusr/test")
	require.NoError(t, err)
	require.NotEqual(t, key, dev.ApiKey)
	dev, err = o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	require.Equal(t, key, dev.ApiKey)

	devs, err := db.ReadAllDevices("tstusr")
	require.NoError(t, err)
	require.Equal(t, 3, len(devs)) //All devices
	devs, err = o.ReadAllDevices("tstusr")
	require.NoError(t, err)
	require.Equal(t, 1, len(devs)) //Only this device

	usr, err := db.ReadUser("tstusr")
	require.NoError(t, err)
	devs, err = o.ReadAllDevicesByUserID(usr.UserId)
	require.NoError(t, err)
	require.Equal(t, 1, len(devs))

	require.Error(t, o.DeleteDevice("tstusr/test"))

	//Now make it an admin device
	require.NoError(t, db.SetAdmin("tstusr/test", true))
	require.NoError(t, o.SetAdmin("tstusr/testdevice", true))

	require.NoError(t, o.DeleteDevice("tstusr/testdevice"))

}
