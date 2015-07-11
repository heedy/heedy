package streamdb

import (
	"testing"

	"connectordb/config"

	"github.com/stretchr/testify/require"
)

func TestDatabaseStreamCrud(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()

	require.NoError(t, db.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tst/testdevice"))

	strms, err := db.ReadAllStreams("tst/testdevice")
	require.NoError(t, err)
	require.Equal(t, 0, len(strms))

	require.Error(t, db.CreateStream("tst/testdevice/stream1", `{"type":"ssafda sdg"}`))

	require.NoError(t, db.CreateStream("tst/testdevice/stream1", `{"type":"string"}`))

	strms, err = db.ReadAllStreams("tst/testdevice")
	require.NoError(t, err)
	require.Equal(t, 1, len(strms))

	_, err = db.ReadStream("tst/testdevice/nostream")
	require.Error(t, err)

	s, err := db.ReadStream("tst/testdevice/stream1")
	require.NoError(t, err)
	require.Equal(t, "stream1", s.Name)

	s.Name = "stream2"
	require.NoError(t, db.UpdateStream(s))

	_, err = db.ReadStream("tst/testdevice/stream1")
	require.Error(t, err)

	s, err = db.ReadStream("tst/testdevice/stream2")
	require.NoError(t, err)
	require.Equal(t, "stream2", s.Name)

	require.Equal(t, "string", s.Schema["type"])

	s.StreamId = 3634
	require.Error(t, db.UpdateStream(s))

	require.Error(t, db.DeleteStream("tst/testdevice/stream1"))
	require.NoError(t, db.DeleteStream("tst/testdevice/stream2"))

	require.NoError(t, db.CreateStream("tst/testdevice/stream1", `{"type":"string"}`))
	require.NoError(t, db.CreateStream("tst/testdevice/stream2", `{"type":"string"}`))
	require.NoError(t, db.CreateStream("tst/testdevice/stream3", `{"type":"string"}`))

	strms, err = db.ReadAllStreams("tst/testdevice")
	require.NoError(t, err)
	require.Equal(t, 3, len(strms))

	strm, err := db.ReadStream("tst/testdevice/stream1")
	require.NoError(t, err)

	db.Reload()
	strm2, err := db.ReadStreamByID(strm.StreamId)
	require.NoError(t, err)
	require.Equal(t, strm2.Name, strm.Name)

	strm, err = db.ReadStreamByDeviceID(strm.DeviceId, "stream1")
	require.NoError(t, err)
	require.Equal(t, strm.Name, strm2.Name)

}
