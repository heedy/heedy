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

}
