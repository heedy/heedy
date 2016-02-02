package connectordb

import (
	"connectordb/datastream"
	"connectordb/messenger"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ensureUserlog(t *testing.T, m messenger.Message, cmd, arg string) {
	require.Equal(t, "streamdb_test/meta/log", m.Stream, cmd)

	d, ok := m.Data[0].Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, cmd, d["cmd"])
	require.Equal(t, arg, d["arg"])
}

func TestUserlog(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass", "admin", true))

	o, err := db.AsUser("streamdb_test")
	require.NoError(t, err)

	//Now subscribe to the userlog
	recvchan := make(chan messenger.Message, 2)
	_, err = o.Subscribe("streamdb_test/meta/log", recvchan)
	require.NoError(t, err)

	//The message timeout
	go func() {
		time.Sleep(5 * time.Second)
		recvchan <- messenger.Message{"TIMEOUT", "", []datastream.Datapoint{}}
	}()

	o.CreateDevice("streamdb_test/mydevice")
	ensureUserlog(t, <-recvchan, "CreateDevice", "streamdb_test/mydevice")

	o.UpdateDevice("streamdb_test/mydevice", map[string]interface{}{"nickname": "hiah"})
	ensureUserlog(t, <-recvchan, "UpdateDevice", "streamdb_test/mydevice")

	o.CreateStream("streamdb_test/mydevice/mystream", "{\"type\": \"string\"}")
	ensureUserlog(t, <-recvchan, "CreateStream", "streamdb_test/mydevice/mystream")

	{
		err = o.UpdateStream("streamdb_test/mydevice/mystream", map[string]interface{}{"nickname": "hiah"})
		require.NoError(t, err)
		ensureUserlog(t, <-recvchan, "UpdateStream", "streamdb_test/mydevice/mystream")
	}

	err = o.DeleteStream("streamdb_test/mydevice/mystream")
	require.NoError(t, err)
	ensureUserlog(t, <-recvchan, "DeleteStream", "streamdb_test/mydevice/mystream")

	require.NoError(t, o.DeleteDevice("streamdb_test/mydevice"))
	ensureUserlog(t, <-recvchan, "DeleteDevice", "streamdb_test/mydevice")

	require.NoError(t, o.UpdateUser("streamdb_test", map[string]interface{}{"email": "hey@localhost"}))
	ensureUserlog(t, <-recvchan, "UpdateUser", "streamdb_test")

	require.NoError(t, db.UpdateUser("streamdb_test", map[string]interface{}{"role": "admin"}))

	require.NoError(t, o.CreateUser("starry_eyed_userlog", "rofl@localhost", "mypass", "user", true))
	ensureUserlog(t, <-recvchan, "CreateUser", "starry_eyed_userlog")

	require.NoError(t, o.DeleteUser("starry_eyed_userlog"))
	ensureUserlog(t, <-recvchan, "DeleteUser", "starry_eyed_userlog")

}
