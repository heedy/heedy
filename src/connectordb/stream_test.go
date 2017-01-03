package connectordb

import (
	"connectordb/users"
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

	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "myuser", Email: "email@email", Password: "test", Role: "user", Public: true}}))
	require.NoError(t, db.CreateDevice("myuser/mydevice", &users.DeviceMaker{}))

	require.Error(t, db.CreateStream("nouser/mydevice/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type":"number"}`}}))
	require.Error(t, db.CreateStream("myuser/nodevice/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type":"number"}`}}))
	require.Error(t, db.CreateStream("myuser/mydevice/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type":"foobar"}`}}))
	require.NoError(t, db.CreateStream("myuser/mydevice/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type":"number"}`}}))

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

	require.NoError(t, db.CreateStream("myuser/mydevice/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type":"number"}`}}))
	u, err = db.ReadStream("myuser/mydevice/mystream")
	require.NoError(t, err)
	require.NoError(t, db.DeleteUser("myuser"))
	_, err = db.ReadStreamByID(u.StreamID)
	require.Error(t, err)
}

func TestStreamUpdate(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "myuser", Email: "email@email", Password: "test", Role: "user", Public: true}}))
	require.NoError(t, db.CreateDevice("myuser/mydevice", &users.DeviceMaker{}))
	require.NoError(t, db.CreateStream("myuser/mydevice/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type":"number"}`}}))

	require.Error(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"name": "lol"}))
	require.Error(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"schema": `{"type": "str    ing"}`}))
	require.Error(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"foobar": "blah"}))

	require.NoError(t, db.UpdateStream("myuser/mydevice/mystream", map[string]interface{}{"nickname": "hi"}))

	u, err := db.ReadStream("myuser/mydevice/mystream")
	require.NoError(t, err)
	require.Equal(t, "hi", u.Nickname)

}

func TestReadUserStreams(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "myuser", Email: "email@email", Password: "test", Role: "user", Public: true}}))
	require.NoError(t, db.CreateDevice("myuser/mydevice", &users.DeviceMaker{Device: users.Device{IsVisible: true}}))
	require.NoError(t, db.CreateStream("myuser/mydevice/mystream", &users.StreamMaker{}))

	require.NoError(t, db.CreateDevice("myuser/mydevice2", &users.DeviceMaker{}))
	require.NoError(t, db.CreateStream("myuser/mydevice2/mystream", &users.StreamMaker{Stream: users.Stream{Downlink: true}}))

	require.NoError(t, db.CreateDevice("myuser/mydevice3", &users.DeviceMaker{Device: users.Device{Public: true}}))
	require.NoError(t, db.CreateStream("myuser/mydevice3/mystream", &users.StreamMaker{Stream: users.Stream{Downlink: true}}))

	s, err := db.ReadUserStreams("myuser", false, false, false)
	require.NoError(t, err)
	require.Equal(t, 3, len(s))

	s, err = db.ReadUserStreams("myuser", true, false, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(s))

	s, err = db.ReadUserStreams("myuser", false, true, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(s))

	s, err = db.ReadUserStreams("myuser", false, false, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(s))

	s, err = db.ReadUserStreams("myuser", true, true, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(s))

}
