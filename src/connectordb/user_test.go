package connectordb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	num, err := db.CountUsers()
	require.NoError(t, err)
	require.Equal(t, int64(0), num)

	u, err := db.ReadUser("tst")
	require.Nil(t, u)
	require.Error(t, err)

	require.Error(t, db.CreateUser("myuser", "email@email", "mypass", "notarole", true))
	require.Error(t, db.CreateUser("myuser", "not an email", "mypass", "user", true))
	require.Error(t, db.CreateUser("myuser", "email@email", "", "user", true))

	require.NoError(t, db.CreateUser("myuser", "email@email", "test", "user", true))
	require.Error(t, db.CreateUser("myuser", "email@email", "test2", "user", true))

	u, err = db.ReadUser("myuser")
	require.NoError(t, err)
	require.Equal(t, "myuser", u.Name)
	require.Equal(t, "email@email", u.Email)
	require.Equal(t, "user", u.Role)

	require.NoError(t, db.DeleteUser("myuser"))
	require.Error(t, db.DeleteUser("myuser"))

	_, err = db.ReadUser("tst")
	require.Error(t, err)

	require.NoError(t, db.CreateUser("myuser", "email@email", "test", "user", true))
}

func TestUserUpdate(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	require.NoError(t, db.CreateUser("myuser", "email@email", "test", "user", true))

	require.Error(t, db.UpdateUser("myuser", map[string]interface{}{"name": "lol"}))
	require.Error(t, db.UpdateUser("myuser", map[string]interface{}{"role": "rawr"}))
	require.Error(t, db.UpdateUser("myuser", map[string]interface{}{"foobar": "blah"}))

	require.NoError(t, db.UpdateUser("myuser", map[string]interface{}{"nickname": "hi"}))

	u, err := db.ReadUser("myuser")
	require.NoError(t, err)
	require.Equal(t, "hi", u.Nickname)
}
