package streamdb

import (
	"testing"

	"connectordb/config"

	"github.com/stretchr/testify/require"
)

func TestAuthStreamCrud(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()

	_, err = db.ReadAllStreams("bad/badder")
	require.Error(t, err)

	require.NoError(t, db.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tst/testdevice"))

	require.NoError(t, db.CreateDevice("tst/testdevice2"))
	require.NoError(t, db.CreateStream("tst/testdevice2/teststream", `{"type":"string"}`))

	o, err := db.GetOperator("tst/testdevice")
	require.NoError(t, err)

	dev, err := db.ReadDevice("tst/testdevice2")
	require.NoError(t, err)
	_, err = o.ReadAllStreamsByDeviceID(dev.DeviceId)
	require.Error(t, err)

	_, err = o.ReadAllStreams("tst/testdevice2")
	require.Error(t, err)

	dev, err = o.Device()
	require.NoError(t, err)
	strms, err := o.ReadAllStreamsByDeviceID(dev.DeviceId)
	require.NoError(t, err)
	require.Equal(t, 0, len(strms))

	strms, err = o.ReadAllStreams("tst/testdevice")
	require.NoError(t, err)
	require.Equal(t, 0, len(strms))

	require.Error(t, o.CreateStream("tst/testdevice2/mystream", `{"type":"string"}`))
	require.NoError(t, o.CreateStream("tst/testdevice/mystream", `{"type":"string"}`))

	_, err = o.ReadStream("tst/testdevice2/teststream")
	require.Error(t, err)

	s, err := o.ReadStream("tst/testdevice/mystream")
	require.NoError(t, err)
	require.Equal(t, "mystream", s.Name)

	s.Name = "stream2"
	require.NoError(t, o.UpdateStream(s))

	_, err = o.ReadStream("tst/testdevice/mystream")
	require.Error(t, err)

	s, err = db.ReadStream("tst/testdevice/stream2")
	require.NoError(t, err)
	require.Equal(t, "stream2", s.Name)

	require.Error(t, o.DeleteStream("tst/testdevice2/teststream"))
	require.NoError(t, o.DeleteStream("tst/testdevice/stream2"))

	_, err = db.ReadStream("tst/testdevice/mystream")
	require.Error(t, err)

	dev, err = o.ReadDevice("tst/testdevice")
	require.NoError(t, err)

	require.NoError(t, o.CreateStreamByDeviceID(dev.DeviceId, "testme", `{"type":"string"}`))

	s, err = o.ReadStreamByDeviceID(dev.DeviceId, "testme")
	require.NoError(t, err)
	require.Equal(t, s.Name, "testme")
	require.NoError(t, o.DeleteStreamByID(s.StreamId, ""))
	_, err = o.ReadStreamByID(s.StreamId)
	require.Error(t, err)
	_, err = o.ReadStreamByDeviceID(dev.DeviceId, "testme")
	require.Error(t, err)
}
