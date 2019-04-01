package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserUser(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	// Create
	name := "testy"
	require.NoError(t, adb.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

	db := NewUserDB(adb, "testy")

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
	adb.AddUserScopes("testy", "users:create")

	// Create
	name2 := "testy2"
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name2,
			},
		},
		Password: "testpass",
	}))

	_, err := db.ReadUser("notauser", false)
	require.Error(t, err)

	u, err := db.ReadUser("testy", false)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	_, err = db.ReadUser("testy2", false)
	require.Error(t, err)

	// Make sure we can no longer read ourselves if we remove the wrong permission
	adb.RemUserScopes("testy", "user:read")
	adb.RemGroupScopes("users", "user:read")

	_, err = db.ReadUser("testy", false)
	require.Error(t, err)

	adb.AddGroupScopes("users", "users:read")

	u, err = db.ReadUser("testy", false)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	u, err = db.ReadUser("testy2", false)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy2")

	require.NoError(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: "testy",
			},
		},
		Password: "mypass2",
	}))

	// Shouldn't be allowed to change another user's password without the scope present
	require.Error(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: "testy2",
			},
		},
		Password: "mypass2",
	}))

	adb.AddUserScopes("testy", "users:edit:password")

	require.NoError(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: "testy2",
			},
		},
		Password: "mypass2",
	}))

	// And now try deleting the user
	require.Error(t, db.DelUser("testy2"))

	adb.AddGroupScopes("testy", "user:delete")

	require.Error(t, db.DelUser("testy2"))

	adb.AddGroupScopes("users", "users:delete")

	require.NoError(t, db.DelUser("testy2"))

	_, err = adb.ReadUser("testy2", false)
	require.Error(t, err)

	require.NoError(t, db.DelUser("testy"))

	// And now comes the question of ensuring that the db object is no longer valid...
	// but a user only logs in from browser, so maybe can just manually check eveny couple minutes?

}

func TestUserScopes(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	// Create
	name := "testy"
	require.NoError(t, adb.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

	name2 := "testy2"
	require.NoError(t, adb.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name2,
			},
		},
		Password: "testpass",
	}))

	db := NewUserDB(adb, "testy")

	s, err := db.GetUserScopes("testy")
	require.NoError(t, err)

	require.NoError(t, adb.AddGroupScopes("users", "trunk"))
	require.NoError(t, adb.AddUserScopes("testy", "retfdg"))

	s2, err := db.GetUserScopes("testy")
	require.NoError(t, err)
	require.Equal(t, len(s)+2, len(s2))

	_, err = db.GetUserScopes("testy2")
	require.Error(t, err)

	require.NoError(t, adb.AddGroupScopes("users", "users:scopes"))
	s, err = db.GetUserScopes("testy2")
	require.NoError(t, err)

	require.NoError(t, adb.AddGroupScopes("users", "trueertwertnk"))
	require.NoError(t, adb.AddUserScopes("testy2", "retfgshfdgaerdg"))

	s2, err = db.GetUserScopes("testy2")
	require.NoError(t, err)
	require.Equal(t, len(s)+2, len(s2))
}
