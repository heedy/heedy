/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package authoperator_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthUserCrud(t *testing.T) {
	db.Clear()

	//Create extra users that exist
	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass", "user", true))
	require.NoError(t, db.CreateUser("streamdb_test2", "root@localhost2", "mypass", "admin", false))
	require.NoError(t, db.CreateUser("streamdb_test3", "root@localhost3", "mypass", "admin", false))

	o, err := db.AsUser("streamdb_test")
	require.NoError(t, err)

	// Try to create a user not as an admin
	require.Error(t, o.CreateUser("notanadmin", "lol@you", "fail", "user", true))

	require.Error(t, o.UpdateUser("streamdb_test", map[string]interface{}{"role": "admin"}))

	//Make sure there are 3
	usrs, err := db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 3, len(usrs))

	//Now make sure that auth is working correctly
	_, err = o.ReadAllUsers()
	require.Error(t, err)
	//require.Equal(t, 1, len(usrs))
	//require.Equal(t, "streamdb_test", usrs[0].Name)

	u, err := o.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	u, err = o.ReadUser("streamdb_test2")
	require.Error(t, err)
	u, err = o.ReadUser("notauser")
	require.Error(t, err)

	require.NoError(t, o.UpdateUser("streamdb_test", map[string]interface{}{"password": "pass2"}))

	_, err = db.UserLogin("streamdb_test", "pass2")
	require.NoError(t, err)

	u, err = o.User()
	require.NoError(t, err)

	require.Error(t, o.DeleteUser("streamdb_test2"))
	require.Error(t, o.DeleteUser("streamdb_test"))

	//Now, let's make this an admin user
	require.NoError(t, db.UpdateUser("streamdb_test", map[string]interface{}{"role": "admin"}))

	u, err = db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "admin", u.Role)

	//Make sure there are 3 if admin
	usrs, err = o.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 3, len(usrs))

	u, err = o.ReadUser("streamdb_test2")
	require.NoError(t, err)

	require.NoError(t, o.DeleteUser("streamdb_test2"))

	_, err = db.AsUser("streamdb_test2")
	require.Error(t, err)

	o, err = db.AsUser("streamdb_test3")
	require.NoError(t, err)

	require.NoError(t, o.UpdateUser("streamdb_test3", map[string]interface{}{"role": "user", "public": true}))

	u, err = o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test3", u.Name)
	require.Error(t, o.DeleteUserByID(u.UserID))

	require.NoError(t, db.UpdateUser("streamdb_test3", map[string]interface{}{"role": "admin"}))
	require.NoError(t, o.DeleteUserByID(u.UserID))
	_, err = o.User()
	require.Error(t, err)
}

func TestNobodyUser(t *testing.T) {
	db.Clear()

	//Create extra users that exist
	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass", "user", true))

	n := db.Nobody()

	_, err := n.ReadUser("streamdb_test")
	require.Error(t, err)
}
