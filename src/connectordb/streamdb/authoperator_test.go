package streamdb

import (
	"testing"

	"connectordb/config"

	"github.com/stretchr/testify/require"
)

func TestAuthOperatorBasics(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	_, err = db.UserLoginOperator("streamdb_test", "wrongpass")
	require.Error(t, err)

	o, err := db.UserLoginOperator("streamdb_test", "mypass")
	require.NoError(t, err)

	require.Equal(t, "streamdb_test/user", o.Name())

	u, err := o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	d, err := o.Device()
	require.NoError(t, err)
	require.Equal(t, "user", d.Name)

	o, err = db.APILoginOperator(d.ApiKey)
	require.NoError(t, err)
	require.Equal(t, "streamdb_test/user", o.Name())

}

func TestCountAll(t *testing.T) {
	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	o, err := db.UserLoginOperator("streamdb_test", "mypass")
	require.NoError(t, err)

	_, err = o.CountAllUsers()
	require.Error(t, err)
	_, err = o.CountAllDevices()
	require.Error(t, err)
	_, err = o.CountAllStreams()
	require.Error(t, err)

	db.SetAdmin("streamdb_test", true)

	i, err := o.CountAllUsers()
	require.NoError(t, err)
	require.EqualValues(t, i, 1)
	i, err = o.CountAllDevices()
	require.NoError(t, err)
	require.True(t, i >= 1)
	i, err = o.CountAllStreams()
	require.NoError(t, err)
	require.True(t, i >= 1)
}
