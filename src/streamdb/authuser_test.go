package streamdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthUserCrud(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()
	//go db.RunWriter()

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	_, err = db.UserLoginOperator("streamdb_test", "wrongpass")
	require.Error(t, err)

	o, err := db.UserLoginOperator("streamdb_test", "mypass")
	require.NoError(t, err)

	require.Equal(t, "streamdb_test/user", o.Name())
	require.Equal(t, db, o.Database())

	u, err := o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	d, err := o.Device()
	require.NoError(t, err)
	require.Equal(t, "user", d.Name)

	require.Error(t, o.CreateUser("notanadmin", "lol@you", "fail"))

	//Create extra users that exist
	require.NoError(t, db.CreateUser("streamdb_test2", "root@localhost2", "mypass"))
	require.NoError(t, db.CreateUser("streamdb_test3", "root@localhost3", "mypass"))

	//Make sure there are 3
	usrs, err := db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 3, len(usrs))

	//Now make sure that auth is working correctly
	usrs, err = o.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 1, len(usrs))
	require.Equal(t, "streamdb_test", usrs[0].Name)

	u, err = o.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	u, err = o.ReadUser("streamdb_test2")
	require.Error(t, err)
	u, err = o.ReadUser("notauser")
	require.Error(t, err)

	require.Error(t, o.SetAdmin("streamdb_test", true))

	require.NoError(t, o.ChangeUserPassword("streamdb_test", "pass2"))

	_, err = db.UserLoginOperator("streamdb_test", "pass2")
	require.NoError(t, err)

	u, err = o.ReadUserByEmail("root@localhost2")
	require.Error(t, err)

	u, err = o.ReadUserByEmail("root@localhost")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	u.Email = "testemail@test.com"
	require.NoError(t, o.UpdateUser(u.Name, u))

	u, err = o.User()
	require.NoError(t, err)
	require.Equal(t, "testemail@test.com", u.Email)

	require.Error(t, o.DeleteUser("streamdb_test2"))
	require.Error(t, o.DeleteUser("streamdb_test"))

	//Now, let's make this an admin user
	require.NoError(t, db.SetAdmin("streamdb_test", true))

	//Reload the device with admin
	require.NoError(t, o.Reload())

	//Make sure there are 3 if admin
	usrs, err = o.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 3, len(usrs))

	u, err = o.ReadUser("streamdb_test2")
	require.NoError(t, err)

	u, err = o.ReadUserByEmail("root@localhost2")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test2", u.Name)

	require.NoError(t, o.DeleteUser("streamdb_test2"))

	_, err = db.Operator("streamdb_test2")
	require.Error(t, err)
	o, err = db.Operator("streamdb_test3")
	require.NoError(t, err)

	u, err = o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test3", u.Name)

}
