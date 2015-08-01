package authoperator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator/interfaces"
	"connectordb/streamdb/operator/messenger"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ensureUserlog(t *testing.T, m messenger.Message, cmd, arg string) {
	require.Equal(t, "streamdb_test/user/log", m.Stream)
	require.Equal(t, "streamdb_test/user", m.Data[0].Sender)

	d, ok := m.Data[0].Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, cmd, d["cmd"])
	require.Equal(t, arg, d["arg"])
}

func TestUserlog(t *testing.T) {
	fmt.Println("test userlog")

	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	require.NoError(t, baseOperator.CreateUser("streamdb_test", "root@localhost", "mypass"))

	ao, err := NewUserAuthOperator(baseOperator, "streamdb_test")
	require.NoError(t, err)
	o := interfaces.PathOperatorMixin{ao}

	//Now subscribe to the userlog
	recvchan := make(chan messenger.Message, 2)
	_, err = o.Subscribe("streamdb_test/user/log", recvchan)
	require.NoError(t, err)

	//The message timeout
	go func() {
		time.Sleep(5 * time.Second)
		recvchan <- messenger.Message{"TIMEOUT", []datastream.Datapoint{}}
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

	{
		s, err := o.ReadStream("streamdb_test/mydevice/mystream")
		require.NoError(t, err)
		s.Nickname = "hiah"
		err = o.UpdateStream(s)
		require.NoError(t, err)
		ensureUserlog(t, <-recvchan, "UpdateStream", "streamdb_test/mydevice/mystream")
	}

	o.DeleteStream("streamdb_test/mydevice/mystream")
	ensureUserlog(t, <-recvchan, "DeleteStream", "streamdb_test/mydevice/mystream")

	o.DeleteDevice("streamdb_test/mydevice")
	ensureUserlog(t, <-recvchan, "DeleteDevice", "streamdb_test/mydevice")

	usr, err := o.ReadUser("streamdb_test")

	usr.Email = "hey@localhost"
	o.UpdateUser(usr)
	ensureUserlog(t, <-recvchan, "UpdateUser", "streamdb_test")

	baseOperator.SetAdmin("streamdb_test", true)

	o.CreateUser("starry_eyed_userlog", "rofl@localhost", "mypass")
	ensureUserlog(t, <-recvchan, "CreateUser", "starry_eyed_userlog")

	o.DeleteUser("starry_eyed_userlog")
	ensureUserlog(t, <-recvchan, "DeleteUser", "starry_eyed_userlog")

}
