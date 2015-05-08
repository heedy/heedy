package streamdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseStreamCrud(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()

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
	require.NoError(t, db.UpdateStream("tst/testdevice/stream1", s))

	_, err = db.ReadStream("tst/testdevice/stream1")
	require.Error(t, err)

	s, err = db.ReadStream("tst/testdevice/stream2")
	require.NoError(t, err)
	require.Equal(t, "stream2", s.Name)

	require.Equal(t, "string", s.Schema["type"])

	s.StreamId = 3634
	require.Error(t, db.UpdateStream("tst/testdevice/stream2", s))

	require.Error(t, db.DeleteStream("tst/testdevice/stream1"))
	require.NoError(t, db.DeleteStream("tst/testdevice/stream2"))

	require.NoError(t, db.CreateStream("tst/testdevice/stream1", `{"type":"string"}`))
	require.NoError(t, db.CreateStream("tst/testdevice/stream2", `{"type":"string"}`))
	require.NoError(t, db.CreateStream("tst/testdevice/stream3", `{"type":"string"}`))

	strms, err = db.ReadAllStreams("tst/testdevice")
	require.NoError(t, err)
	require.Equal(t, 3, len(strms))

	_, err = db.ReadStream("tst/testdevice/stream1")
	require.NoError(t, err)

	require.NoError(t, db.DeleteDeviceStreams("tst/testdevice"))

	_, err = db.ReadStream("tst/testdevice/stream1")
	require.Error(t, err)
}
