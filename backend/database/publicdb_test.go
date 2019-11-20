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

	// Can't create a user
	require.Error(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	// Create the user with admin db
	require.NoError(t, adb.CreateUser(&User{
		UserName: &name,
		Password: &passwd,
	}))

	// Can't read the user
	_, err := db.ReadUser("testy", nil)
	require.Error(t, err)

	pread := true
	require.NoError(t, adb.UpdateUser(&User{
		Details: Details{
			ID: name,
		},
		PublicRead: &pread,
	}))

	// But when the user is public, we can

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.UserName, "testy")

	// Modifying a user is no-go
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: &passwd,
	}))

	require.Error(t, db.DelUser("testy"))
}

func TestPublicObject(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	pdb := NewPublicDB(adb)
	name := "tree"
	stype := "stream"
	_, err := pdb.CreateObject(&Object{
		Details: Details{
			Name: &name,
		},
		Type: &stype,
	})
	require.Error(t, err)

	udb := NewUserDB(adb, "testy")
	sid, err := udb.CreateObject(&Object{
		Details: Details{
			Name: &name,
		},
		Type: &stype,
	})
	require.NoError(t, err)

	_, err = pdb.ReadObject(sid, nil)
	require.Error(t, err)

	// Now share the object with public
	require.NoError(t, udb.ShareObject(sid, "public", &ScopeArray{
		Scopes: []string{"read"},
	}))

	s, err := pdb.ReadObject(sid, nil)
	require.NoError(t, err)

	require.Equal(t, *s.Details.Name, name)

}
