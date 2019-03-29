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

	// Can't create the user
	require.EqualError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}), ErrAccessDenied.Error())

	// Add user creation permission
	adb.AddGroupScopes("public", "users:create")

	// Create
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))
	_, err := db.ReadUser("testy")
	require.Error(t, err)

	adb.AddGroupScopes("public", "users:read")

	u, err := db.ReadUser("testy")
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	// Shouldn't be allowed to change another user's password without the scope present
	require.Error(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: "testy",
			},
		},
		Password: "mypass2",
	}))

	adb.AddGroupScopes("public", "users:edit:password")

	require.NoError(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: "testy",
			},
		},
		Password: "mypass2",
	}))

	require.Error(t, db.DelUser("testy"))
	adb.AddGroupScopes("public", "users:delete")
	require.NoError(t, db.DelUser("testy"))

	_, err = adb.ReadUser("testy")
	require.Error(t, err)
}
