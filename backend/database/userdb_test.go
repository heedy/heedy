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
	passwd := "testpass"
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	db := NewUserDB(adb, "testy")

	// Can't create the user
	require.Error(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	// Add user creation permission
	adb.AddUserScopeSet("testy", "testy")
	adb.AddScope("testy", "users:create")

	// Create
	name2 := "testy2"
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name2,
		},
		Password: &passwd,
	}))

	_, err := db.ReadUser("notauser", nil)
	require.Error(t, err)

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	_, err = db.ReadUser("testy2", nil)
	require.Error(t, err)

	adb.AddScope("users", "users:read")

	u, err = db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	u, err = db.ReadUser("testy2", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy2")

	passwd = "mypass2"
	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: &passwd,
	}))

	// Shouldn't be allowed to change another user's password without the scope present
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy2",
		},
		Password: &passwd,
	}))

	adb.AddScope("testy", "users:edit", "users:edit:password")

	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy2",
		},
		Password: &passwd,
	}))

	// And now try deleting the user
	require.Error(t, db.DelUser("testy2"))

	adb.AddScope("testy", "user:delete")

	require.Error(t, db.DelUser("testy2"))

	adb.AddScope("users", "users:delete")

	require.NoError(t, db.DelUser("testy2"))

	_, err = adb.ReadUser("testy2", nil)
	require.Error(t, err)

	require.NoError(t, db.DelUser("testy"))

	// And now comes the question of ensuring that the db object is no longer valid...
	// but a user only logs in from browser, so maybe can just manually check eveny couple minutes?

}

func TestUserUpdateAvatar(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	db := NewUserDB(adb, "testy")
	avatar := "mi:lol"
	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID:     "testy",
			Avatar: &avatar,
		},
	}))
}

func TestUserScopes(t *testing.T) {
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

	name2 := "testy2"
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name2,
		},
		Password: &passwd,
	}))

	db := NewUserDB(adb, "testy")

	s, err := db.ReadUserScopes("testy")
	require.NoError(t, err)

	require.NoError(t, adb.AddUserScopeSet("testy", "testy"))
	require.NoError(t, adb.AddScope("users", "trunk"))
	require.NoError(t, adb.AddScope("testy", "retfdg"))

	s2, err := db.ReadUserScopes("testy")
	require.NoError(t, err)
	require.Equal(t, len(s)+2, len(s2))

	_, err = db.ReadUserScopes("testy2")
	require.Error(t, err)

	require.NoError(t, adb.AddScope("users", "users:scopes"))
	uacc := 100
	require.NoError(t, adb.UpdateUser(&User{
		Details: Details{
			ID: "testy2",
		},
		UserAccess: &uacc,
	}))
	s, err = db.ReadUserScopes("testy2")
	require.NoError(t, err)
	require.Equal(t, len(s), 2)

	require.NoError(t, adb.AddScope("users", "trueertwertnk"))
	require.NoError(t, adb.AddScope("testy2", "retfgshfdgaerdg"))
	require.NoError(t, adb.AddUserScopeSet("testy2", "testy2"))

	s2, err = db.ReadUserScopes("testy2")
	require.NoError(t, err)
	require.Equal(t, len(s)+2, len(s2))
}

func TestUserStreams(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	name := "testy"
	name2 := "testy2"
	passwd := "testpass"
	canRead := 100
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name2,
		},
		Password:   &passwd,
		UserAccess: &canRead,
	}))

	db := NewUserDB(adb, "testy")
	sname := "streamy"
	_, err := db.CreateStream(&Stream{
		Details: Details{
			Name: &sname,
		},
		Owner: &name2,
	})
	require.Error(t, err)

	s1, err := db.CreateStream(&Stream{
		Details: Details{
			Name: &sname,
		},
		Owner: &name,
	})
	require.NoError(t, err)

	adb.AddScope("users", "streams:create")

	s2, err := db.CreateStream(&Stream{
		Details: Details{
			Name: &sname,
		},
		Owner: &name2,
	})
	require.NoError(t, err)

	s, err := db.ReadStream(s1, nil)
	require.NoError(t, err)
	require.Equal(t, s.ID, s1)

	_, err = db.ReadStream(s2, nil)
	require.Error(t, err)

	adb.AddScope("public", "streams:read")
	s, err = db.ReadStream(s2, nil)
	require.NoError(t, err)
	require.Equal(t, s.ID, s2)

	fname := "booya"
	require.NoError(t, db.UpdateStream(&Stream{
		Details: Details{
			ID:       s1,
			FullName: &fname,
		},
	}))

	s, err = db.ReadStream(s1, nil)
	require.NoError(t, err)
	require.NotNil(t, s.FullName)
	require.Equal(t, *s.FullName, fname)

	require.NoError(t, db.DelStream(s1))

	require.Error(t, db.UpdateStream(&Stream{
		Details: Details{
			ID:       s2,
			FullName: &fname,
		},
	}))

	adb.AddScope("users", "streams:edit")
	require.NoError(t, db.UpdateStream(&Stream{
		Details: Details{
			ID:       s2,
			FullName: &fname,
		},
	}))

	s, err = db.ReadStream(s2, nil)
	require.NoError(t, err)
	require.NotNil(t, s.FullName)
	require.Equal(t, *s.FullName, fname)

	require.Error(t, db.DelStream(s2))

	adb.AddScope("users", "streams:delete")
	require.NoError(t, db.DelStream(s2))
}
