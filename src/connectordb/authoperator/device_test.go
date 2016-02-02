package authoperator_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthDeviceUserCrud(t *testing.T) {
	db.Clear()
	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass", "user", true))
	require.NoError(t, db.CreateUser("otheruser", "root@localhost2", "mypass", "admin", false))
	require.NoError(t, db.CreateDevice("otheruser/testdevice"))
	_, err := db.ReadDevice("otheruser/testdevice")
	require.NoError(t, err)

	o, err := db.AsUser("streamdb_test")
	require.NoError(t, err)

	devs, err := o.ReadUserDevices("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, 2, len(devs)) //the user and meta devices

	dev, err := o.Device()
	require.NoError(t, err)

	o2, err := db.AsDevice("streamdb_test/user")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test/user", o2.Name())

	devs, err = o.ReadAllDevicesByUserID(dev.UserID)
	require.NoError(t, err)
	require.Equal(t, 2, len(devs)) //the user and meta device

	// This user should not be able to CRUD devices of another user
	devs, err = o.ReadUserDevices("otheruser")
	require.Error(t, err)

	devs, err = o.ReadAllDevicesByUserID(-1)
	require.Error(t, err)

	_, err = o.ReadDevice("otheruser/testdevice")
	require.Error(t, err)
	require.Error(t, o.DeleteDevice("otheruser/testdevice"))
	require.Error(t, o.CreateDevice("otheruser/testdevice2"))
	_, err = db.ReadDevice("otheruser/testdevice2")
	require.Error(t, err)

	dev, err = db.ReadDevice("otheruser/testdevice")
	require.NoError(t, err)
	_, err = o.ReadDeviceByID(dev.DeviceID)
	require.Error(t, err)
	_, err = o.ReadDeviceByUserID(dev.UserID, "testdevice")
	require.Error(t, err)
	require.Error(t, o.DeleteDeviceByID(dev.DeviceID))
	require.Error(t, o.CreateDeviceByUserID(dev.UserID, "testdevice2"))

	require.Error(t, o.UpdateDevice("otheruser/testdevice", map[string]interface{}{"nickname": "test"}))
	require.Error(t, o.UpdateDevice("otheruser/testdevice", map[string]interface{}{"role": "user"}))
	require.Error(t, o.UpdateDevice("otheruser/testdevice", map[string]interface{}{"apikey": ""}))

	//This user should be able to crud its own devices
	require.NoError(t, o.CreateDevice("streamdb_test/testdevice"))
	dev, err = o.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.Equal(t, "testdevice", dev.Name)
	dev, err = o.ReadDeviceByID(dev.DeviceID)
	require.NoError(t, err)
	require.Equal(t, "testdevice", dev.Name)
	dev, err = o.ReadDeviceByUserID(dev.UserID, "testdevice")
	require.NoError(t, err)
	require.Equal(t, "testdevice", dev.Name)

	_, err = db.DeviceLogin(dev.APIKey)
	require.NoError(t, err)

	oldkey := dev.APIKey
	require.NoError(t, o.UpdateDevice("streamdb_test/testdevice", map[string]interface{}{"apikey": ""}))
	dev, err = o.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.NotEqual(t, oldkey, dev.APIKey)
	require.NotEqual(t, "", dev.APIKey)

	_, err = db.DeviceLogin(oldkey)
	require.Error(t, err)

	require.NoError(t, o.DeleteDevice("streamdb_test/testdevice"))

	usr, err := o.User()
	require.NoError(t, o.CreateDeviceByUserID(usr.UserID, "testdevice"))
	dev, err = o.ReadDevice("streamdb_test/testdevice")
	require.NoError(t, err)
	require.NoError(t, o.DeleteDeviceByID(dev.DeviceID))
}

func TestAuthDeviceDeviceCrud(t *testing.T) {
	db.Clear()
	require.NoError(t, db.CreateUser("tstusr", "root@localhost", "mypass", "user", true))
	require.NoError(t, db.CreateDevice("tstusr/testdevice"))
	require.NoError(t, db.CreateDevice("tstusr/test"))

	o, err := db.AsDevice("tstusr/test")
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
	require.Error(t, o.UpdateDevice("tstusr/testdevice", map[string]interface{}{"nickname": "test"}))
	require.Error(t, o.UpdateDevice("tstusr/testdevice", map[string]interface{}{"role": "user"}))
	require.Error(t, o.UpdateDevice("tstusr/testdevice", map[string]interface{}{"apikey": ""}))

	_, err = o.User()
	require.NoError(t, err)

	require.Error(t, o.UpdateUser("tstusr", map[string]interface{}{"email": "changedemail@lol"}))

	//This device should be able to modify itself
	dev, err := o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	require.Equal(t, "test", dev.Name)

	//Shouldn't have those permissions
	require.Error(t, o.UpdateDevice("tstusr/test", map[string]interface{}{"role": "user"}))

	dev, err = o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	//But changing nickname is fine
	require.NoError(t, o.UpdateDevice("tstusr/test", map[string]interface{}{"nickname": "testnick"}))
	dev, err = o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	require.Equal(t, "testnick", dev.Nickname)
	key := dev.APIKey
	require.NoError(t, o.UpdateDevice("tstusr/test", map[string]interface{}{"apikey": ""}))
	dev, err = o.ReadDevice("tstusr/test")
	require.NoError(t, err)
	require.NotEqual(t, key, dev.APIKey)

	_, err = o.ReadUserDevices("tstusr")
	require.Error(t, err)

	require.NoError(t, db.UpdateDevice("tstusr/test", map[string]interface{}{"role": "reader"}))

	devs, err := db.ReadUserDevices("tstusr")
	require.NoError(t, err)
	require.Equal(t, 4, len(devs)) //All devices
	devs, err = o.ReadUserDevices("tstusr")
	require.NoError(t, err)
	require.Equal(t, 4, len(devs)) //Only this device

	usr, err := db.ReadUser("tstusr")
	require.NoError(t, err)
	devs, err = o.ReadAllDevicesByUserID(usr.UserID)
	require.NoError(t, err)
	require.Equal(t, 4, len(devs))

	require.Error(t, o.DeleteDevice("tstusr/testdevice"))

	//Now make it an admin device
	require.NoError(t, db.UpdateDevice("tstusr/test", map[string]interface{}{"role": "user"}))
	require.NoError(t, o.UpdateDevice("tstusr/testdevice", map[string]interface{}{"role": "user"}))

	require.NoError(t, o.DeleteDevice("tstusr/testdevice"))

}
