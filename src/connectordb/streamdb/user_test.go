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
	_, err = db.ReadUserByID(53)
	require.Error(t, err)

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	usrs, err = db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 1, len(usrs))

	usr, err := db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", usr.Name)

	//Clear the cache
	db.Reload()
	usr, err = db.ReadUserByID(usr.UserId)
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", usr.Name)

	usr.Admin = true
	usr.Email = "testemail@test.com"
	require.NoError(t, db.UpdateUser(usr))

	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, true, usr.Admin)
	require.Equal(t, "testemail@test.com", usr.Email)

	usr.UserId = 9001
	require.Error(t, db.UpdateUser(usr))

	require.NoError(t, db.SetAdmin("streamdb_test", false))
	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, false, usr.Admin)

	require.NoError(t, db.SetAdmin("streamdb_test", true))
	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, true, usr.Admin)

	_, err = db.LoginOperator("streamdb_test", "wrongpass")
	require.Error(t, err)
	_, err = db.LoginOperator("streamdb_test", "mypass")
	require.NoError(t, err)

	require.NoError(t, db.ChangeUserPassword("streamdb_test", "pass2"))
	_, err = db.LoginOperator("streamdb_test", "mypass")
	require.Error(t, err)
	_, err = db.LoginOperator("streamdb_test", "pass2")
	require.NoError(t, err)

	//As of now, this part fails - delete of nonexisting does not error
	require.Error(t, db.DeleteUser("notauser"))
	require.NoError(t, db.DeleteUser("streamdb_test"))

	_, err = db.ReadUser("streamdb_test")
	require.Error(t, err)

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))
	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", usr.Name)

	require.Error(t, db.DeleteUserByID(345))
	require.NoError(t, db.DeleteUserByID(usr.UserId))

	_, err = db.ReadUser("streamdb_test")
	require.Error(t, err)
	_, err = db.ReadUserByID(usr.UserId)
	require.Error(t, err)
}
