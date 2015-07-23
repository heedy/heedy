package streamdb

import (
	"connectordb/config"
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ensureUserlog(t *testing.T, m operator.Message, cmd, arg string) {
	require.Equal(t, "streamdb_test/user/log", m.Stream)
	require.Equal(t, "streamdb_test/user", m.Data[0].Sender)

	d, ok := m.Data[0].Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, cmd, d["cmd"])
	require.Equal(t, arg, d["arg"])
}

func TestUserlog(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	o, err := db.GetOperator("streamdb_test")
	require.NoError(t, err)

	//Now subscribe to the userlog
	recvchan := make(chan operator.Message, 2)
	_, err = o.Subscribe("streamdb_test/user/log", recvchan)
	require.NoError(t, err)

	db.msg.Flush()

	//The message timeout
	go func() {
		time.Sleep(2 * time.Second)
		recvchan <- operator.Message{"TIMEOUT", []datastream.Datapoint{}}
	}()

	o.CreateDevice("streamdb_test/mydevice")
	ensureUserlog(t, <-recvchan, "CreateDevice", "streamdb_test/mydevice")

	d, err := o.ReadDevice("streamdb_test/mydevice")
	require.NoError(t, err)
	d.Nickname = "hiah"
	o.UpdateDevice(d)
	ensureUserlog(t, <-recvchan, "UpdateDevice", "streamdb_test/mydevice")

	o.CreateStream("streamdb_test/mydevice/mystream", "{\"type\": \"string\"}")
	ensureUserlog(t, <-recvchan, "CreateStream", "streamdb_test/mydevice/mystream")

	s, err := o.ReadStream("streamdb_test/mydevice/mystream")
	require.NoError(t, err)
	s.Nickname = "hiah"
	o.UpdateStream(s)
	ensureUserlog(t, <-recvchan, "UpdateStream", "streamdb_test/mydevice/mystream")

	o.DeleteStream("streamdb_test/mydevice/mystream")
	ensureUserlog(t, <-recvchan, "DeleteStream", "streamdb_test/mydevice/mystream")

	o.DeleteDevice("streamdb_test/mydevice")
	ensureUserlog(t, <-recvchan, "DeleteDevice", "streamdb_test/mydevice")

	usr, err := o.ReadUser("streamdb_test")

	usr.Email = "hey@localhost"
	o.UpdateUser(usr)
	ensureUserlog(t, <-recvchan, "UpdateUser", "streamdb_test")

	db.SetAdmin("streamdb_test", true)

	o.CreateUser("starry_eyed_userlog", "rofl@localhost", "mypass")
	ensureUserlog(t, <-recvchan, "CreateUser", "starry_eyed_userlog")

	o.DeleteUser("starry_eyed_userlog")
	ensureUserlog(t, <-recvchan, "DeleteUser", "starry_eyed_userlog")

}
