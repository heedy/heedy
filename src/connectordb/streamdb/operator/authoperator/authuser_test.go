package authoperator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthUserCrud(t *testing.T) {

	// Open and connect to all services.
	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	//Create extra users that exist
	require.NoError(t, baseOperator.CreateUser("streamdb_test", "root@localhost", "mypass"))
	require.NoError(t, baseOperator.CreateUser("streamdb_test2", "root@localhost2", "mypass"))
	require.NoError(t, baseOperator.CreateUser("streamdb_test3", "root@localhost3", "mypass"))

	o, err := NewUserAuthOperator(&baseOperator, "streamdb_test")
	require.NoError(t, err)

	// Try to create a user not as an admin
	require.Error(t, o.CreateUser("notanadmin", "lol@you", "fail"))

	//Make sure there are 3
	usrs, err := baseOperator.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 3, len(usrs))

	//Now make sure that auth is working correctly
	usrs, err = o.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 1, len(usrs))
	require.Equal(t, "streamdb_test", usrs[0].Name)

	u, err := o.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	u, err = o.ReadUser("streamdb_test2")
	require.Error(t, err)
	u, err = o.ReadUser("notauser")
	require.Error(t, err)

	require.Error(t, o.SetAdmin("streamdb_test", true))

	require.NoError(t, o.ChangeUserPassword("streamdb_test", "pass2"))

	_, err = NewUserLoginOperator(&baseOperator, "streamdb_test", "pass2")
	require.NoError(t, err)

	u, err = o.User()
	require.NoError(t, err)

	require.Error(t, o.DeleteUser("streamdb_test2"))
	require.Error(t, o.DeleteUser("streamdb_test"))

	//Now, let's make this an admin user
	require.NoError(t, baseOperator.SetAdmin("streamdb_test", true))

	u, err = baseOperator.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, true, u.Admin)

	//Make sure there are 3 if admin
	usrs, err = o.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 3, len(usrs))

	u, err = o.ReadUser("streamdb_test2")
	require.NoError(t, err)

	require.NoError(t, o.DeleteUser("streamdb_test2"))

	_, err = NewUserAuthOperator(&baseOperator, "streamdb_test2")
	require.Error(t, err)

	o, err = NewUserAuthOperator(&baseOperator, "streamdb_test3")
	require.NoError(t, err)

	u, err = o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test3", u.Name)
	require.Error(t, o.DeleteUserByID(u.UserId))

	require.NoError(t, baseOperator.SetAdmin("streamdb_test3", true))
	require.NoError(t, o.DeleteUserByID(u.UserId))
	_, err = o.User()
	require.Error(t, err)
}
