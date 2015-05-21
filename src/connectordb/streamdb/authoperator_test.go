package streamdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthOperatorBasics(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()

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
