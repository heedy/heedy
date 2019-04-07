package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicUser(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	db := PublicDB{adb}

	name := "testy"
	passwd := "testpass"

	// Can't create the user
	require.Error(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	// Add user creation permission
	adb.AddScope("public", "users:create")

	// Create
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))
	_, err := db.ReadUser("testy", nil)
	require.Error(t, err)

	require.NoError(t, adb.AddScope("public", "users:read"))

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	// Shouldn't be allowed to change another user's password without the scope present
	passwd = "mypass2"
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: &passwd,
	}))

	require.NoError(t, adb.AddScope("public", "users:edit", "users:edit:password"))

	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: &passwd,
	}))

	require.Error(t, db.DelUser("testy"))
	adb.AddScope("public", "users:delete")
	require.NoError(t, db.DelUser("testy"))

	_, err = adb.ReadUser("testy", nil)
	require.Error(t, err)
}

func TestPublicUserScope(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	// Create
	name := "testy"
	passwd := "testpass"
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	db := NewPublicDB(adb)

	_, err := db.ReadUserScopes("testy")
	require.Error(t, err)

	require.NoError(t, adb.AddScope("public", "users:read", "users:scopes"))

	_, err = db.ReadUserScopes("testy")
	require.NoError(t, err)
}
