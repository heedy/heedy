package streamdb

import (
	"database/sql"
	"streamdb/timebatchdb"
	"testing"

	"github.com/stretchr/testify/require"
)

//Testing timebatchdb really messes with everything, so recreate the necessary stuff here
func ResetTimeBatch() error {
	sdb, err := sql.Open("postgres", "postgres://127.0.0.1:52592/connectordb?sslmode=disable")
	if err != nil {
		return err
	}
	defer sdb.Close()
	sdb.Exec("DELETE FROM timebatchtable;")
	sdb.Exec("DELETE FROM Users;")
	sdb.Exec("DELETE FROM Devices;")

	//Clear the redis cache
	rc, err := timebatchdb.OpenRedisCache("localhost:6379", err)
	if err != nil {
		return err
	}
	defer rc.Close()
	rc.Clear()
	return err
}

func TestDatabaseUserCrud(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()
	go db.RunWriter()

	require.Equal(t, db, db.Database())
	require.NoError(t, db.Reload())

	_, err = db.User()
	require.Equal(t, err, ErrAdmin)

	_, err = db.Device()
	require.Equal(t, err, ErrAdmin)

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

	modu := *usr
	modu.Admin = true
	modu.Email = "testemail@test.com"
	require.NoError(t, db.UpdateUser(usr, modu))

	usr, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, true, usr.Admin)
	require.Equal(t, "testemail@test.com", usr.Email)

	modu = *usr
	modu.UserId = 9001
	require.Error(t, db.UpdateUser(usr, modu))

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
