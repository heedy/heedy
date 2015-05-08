package streamdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthStreamCrud(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.ReadAllStreams("bad/badder")
	require.Error(t, err)

	require.NoError(t, db.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tst/testdevice"))

	require.NoError(t, db.CreateDevice("tst/testdevice2"))
	require.NoError(t, db.CreateStream("tst/testdevice2/teststream", `{"type":"string"}`))

	o, err := db.Operator("tst/testdevice")
	require.NoError(t, err)

	_, err = o.ReadAllStreams("tst/testdevice2")
	require.Error(t, err)

	strms, err := o.ReadAllStreams("tst/testdevice")
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
	require.NoError(t, o.UpdateStream("tst/testdevice/mystream", s))

	_, err = o.ReadStream("tst/testdevice/mystream")
	require.Error(t, err)

	s, err = db.ReadStream("tst/testdevice/stream2")
	require.NoError(t, err)
	require.Equal(t, "stream2", s.Name)

	require.Error(t, o.DeleteStream("tst/testdevice2/teststream"))
	require.NoError(t, o.DeleteStream("tst/testdevice/stream2"))

	_, err = db.ReadStream("tst/testdevice/mystream")
	require.Error(t, err)

}
