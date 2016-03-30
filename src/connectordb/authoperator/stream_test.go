package authoperator_test

import (
	"connectordb/users"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthStreamCrud(t *testing.T) {
	db.Clear()
	_, err := db.ReadDeviceStreams("bad/badder")
	require.Error(t, err)

	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "tst", Email: "root@localhost", Password: "mypass", Role: "user", Public: true}}))
	require.NoError(t, db.CreateDevice("tst/testdevice", &users.DeviceMaker{}))

	require.NoError(t, db.CreateDevice("tst/testdevice2", &users.DeviceMaker{}))
	require.NoError(t, db.CreateStream("tst/testdevice2/teststream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "string"}`}}))

	o, err := db.AsDevice("tst/testdevice")
	require.NoError(t, err)

	dev, err := db.ReadDevice("tst/testdevice2")
	require.NoError(t, err)
	_, err = o.ReadAllStreamsByDeviceID(dev.DeviceID)
	require.Error(t, err)

	_, err = o.ReadDeviceStreams("tst/testdevice2")
	require.Error(t, err)

	dev, err = o.Device()
	require.NoError(t, err)
	strms, err := o.ReadAllStreamsByDeviceID(dev.DeviceID)
	require.NoError(t, err)
	require.Equal(t, 0, len(strms))

	strms, err = o.ReadDeviceStreams("tst/testdevice")
	require.NoError(t, err)
	require.Equal(t, 0, len(strms))

	require.Error(t, o.CreateStream("tst/testdevice2/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "string"}`}}))
	require.NoError(t, o.CreateStream("tst/testdevice/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "string"}`}}))

	_, err = o.ReadStream("tst/testdevice2/teststream")
	require.Error(t, err)

	s, err := o.ReadStream("tst/testdevice/mystream")
	require.NoError(t, err)
	require.Equal(t, "mystream", s.Name)

	require.NoError(t, o.UpdateStream("tst/testdevice/mystream", map[string]interface{}{"nickname": "stream2"}))

	s, err = o.ReadStream("tst/testdevice/mystream")

	require.NoError(t, err)
	require.Equal(t, "stream2", s.Nickname)

	require.Error(t, o.DeleteStream("tst/testdevice2/teststream"))
	require.NoError(t, o.DeleteStream("tst/testdevice/mystream"))

	_, err = db.ReadStream("tst/testdevice/mystream")
	require.Error(t, err)

	dev, err = o.ReadDevice("tst/testdevice")
	require.NoError(t, err)

	require.NoError(t, o.CreateStreamByDeviceID(&users.StreamMaker{Stream: users.Stream{DeviceID: dev.DeviceID, Name: "testme", Schema: `{"type": "string"}`}}))

	s, err = o.ReadStreamByDeviceID(dev.DeviceID, "testme")
	require.NoError(t, err)
	require.Equal(t, s.Name, "testme")
	require.NoError(t, o.DeleteStreamByID(s.StreamID, ""))
	_, err = o.ReadStreamByID(s.StreamID)
	require.Error(t, err)
	_, err = o.ReadStreamByDeviceID(dev.DeviceID, "testme")
	require.Error(t, err)
}
