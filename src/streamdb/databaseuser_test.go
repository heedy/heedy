package streamdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseUserCrud(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()
	//go db.RunWriter()

	usrs, err := db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 0, len(usrs))

	_, err = db.ReadUser("streamdb_test")
	require.Error(t, err)
	_, err = db.ReadUserByEmail("root@localhost")
	require.Error(t, err)

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	usrs, err = db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 1, len(usrs))

	usr, err := db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", usr.Name)

	usr, err = db.ReadUserByEmail("root@localhost")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", usr.Name)

	usr.Admin = true
	usr.Email = "testemail@test.com"
	require.NoError(t, db.UpdateUser("streamdb_test", usr))

	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, true, usr.Admin)
	require.Equal(t, "testemail@test.com", usr.Email)

	usr.UserId = 9001
	require.Error(t, db.UpdateUser("streamdb_test", usr))

	require.NoError(t, db.SetAdmin("streamdb_test", false))
	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, false, usr.Admin)

	require.NoError(t, db.SetAdmin("streamdb_test", true))
	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, true, usr.Admin)

	_, err = db.UserLoginOperator("streamdb_test", "wrongpass")
	require.Error(t, err)
	_, err = db.UserLoginOperator("streamdb_test", "mypass")
	require.NoError(t, err)

	require.NoError(t, db.ChangeUserPassword("streamdb_test", "pass2"))
	_, err = db.UserLoginOperator("streamdb_test", "mypass")
	require.Error(t, err)
	_, err = db.UserLoginOperator("streamdb_test", "pass2")
	require.NoError(t, err)

	//As of now, this part fails - delete of nonexisting does not error
	require.Error(t, db.DeleteUser("notauser"))
	require.NoError(t, db.DeleteUser("streamdb_test"))

	_, err = db.ReadUser("streamdb_test")
	require.Error(t, err)
	_, err = db.ReadUserByEmail("streamdb_test")
	require.Error(t, err)

}
