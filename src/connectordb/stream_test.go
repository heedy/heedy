package connectordb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	num, err := db.CountStreams()
	require.NoError(t, err)
	require.Equal(t, int64(0), num)

	u, err := db.ReadStream("tst/tst/tst")
	require.Nil(t, u)
	require.Error(t, err)

	require.NoError(t, db.CreateUser("myuser", "email@email", "test", "user", true))
	require.NoError(t, db.CreateDevice("myuser/mydevice", false))

	require.Error(t, db.CreateStream("nouser/mydevice/mystream", `{"type":"number"}`))
	require.Error(t, db.CreateStream("myuser/nodevice/mystream", `{"type":"number"}`))
	require.Error(t, db.CreateStream("myuser/mydevice/mystream", `{"type":"foobar"}`))
	require.NoError(t, db.CreateStream("myuser/mydevice/mystream", `{"type":"number"}`))

	u, err = db.ReadStream("myuser/mydevice/mystream")
	require.NoError(t, err)
	require.Equal(t, "mystream", u.Name)
	require.Equal(t, `{"type":"number"}`, u.Schema)
	require.Equal(t, false, u.Downlink)
	require.Equal(t, false, u.Ephemeral)

	require.NoError(t, db.DeleteStream("myuser/mydevice/mystream"))

	_, err = db.ReadStream("myuser/mydevice/mystream")
	require.Error(t, err)

	require.Error(t, db.DeleteStream("myuser/mydevice/mystream"))

	require.NoError(t, db.CreateStream("myuser/mydevice/mystream", `{"type":"number"}`))
	u, err = db.ReadStream("myuser/mydevice/mystream")
	require.NoError(t, err)
	require.NoError(t, db.DeleteUser("myuser"))
	_, err = db.ReadStreamByID(u.StreamID)
	require.Error(t, err)
}

func TestStreamUpdate(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	require.NoError(t, db.CreateUser("myuser", "email@email", "test", "user", true))
	require.NoError(t, db.CreateDevice("myuser/mydevice", false))
	require.NoError(t, db.CreateStream("myuser/mydevice/mystream", `{"type": "number"}`))

	require.Error(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"name": "lol"}))
	require.Error(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"schema": `{"type": "string"}`}))
	require.Error(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"foobar": "blah"}))

	require.NoError(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"nickname": "hi"}))

	u, err := db.ReadStream("myuser/mydevice/mystream")
	require.NoError(t, err)
	require.Equal(t, "hi", u.Nickname)

}
