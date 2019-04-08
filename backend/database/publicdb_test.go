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

func TestPublicStreams(t *testing.T) {
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
		Password:     &passwd,
		PublicAccess: &canRead,
	}))

	db := NewPublicDB(adb)
	sname := "streamy"
	_, err := db.CreateStream(&Stream{
		Details: Details{
			Name: &sname,
		},
		Owner: &name2,
	})
	require.Error(t, err)

	adb.AddScope("public", "streams:create")
	s1, err := db.CreateStream(&Stream{
		Details: Details{
			Name: &sname,
		},
		Owner: &name2,
	})
	require.NoError(t, err)

	_, err = db.CreateStream(&Stream{
		Details: Details{
			Name: &sname,
		},
		Owner: &name,
	})
	require.Error(t, err)

	_, err = db.ReadStream(s1, nil)
	require.Error(t, err)

	adb.AddScope("public", "streams:read")
	s, err := db.ReadStream(s1, nil)
	require.NoError(t, err)
	require.Equal(t, s.ID, s1)

	fname := "booya"
	require.Error(t, db.UpdateStream(&Stream{
		Details: Details{
			ID:       s1,
			FullName: &fname,
		},
	}))

	adb.AddScope("public", "streams:edit")
	require.NoError(t, db.UpdateStream(&Stream{
		Details: Details{
			ID:       s1,
			FullName: &fname,
		},
	}))

	require.Error(t, db.DelStream(s1))
	adb.AddScope("public", "streams:delete")
	require.NoError(t, db.DelStream(s1))
}
