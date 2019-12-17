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
		UserName: &name,
		Password: &passwd,
	}))

	db := NewUserDB(adb, "testy")

	// Can't create users
	name2 := "testy2"
	require.Error(t, db.CreateUser(&User{
		UserName: &name2,
		Password: &passwd,
	}))

	// Admin create the user
	require.NoError(t, adb.CreateUser(&User{
		UserName: &name2,
		Password: &passwd,
	}))

	_, err := db.ReadUser("notauser", nil)
	require.Error(t, err)

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.UserName, "testy")

	_, err = db.ReadUser("testy2", nil)
	require.Error(t, err)

	// Add user access to testy2
	uread := true
	require.NoError(t, adb.UpdateUser(&User{
		Details: Details{
			ID: "testy2",
		},
		UsersRead: &uread,
	}))

	u, err = db.ReadUser("testy2", nil)
	require.NoError(t, err)
	require.Equal(t, *u.UserName, "testy2")

	passwd = "mypass2"
	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: &passwd,
	}))

	// Shouldn't be allowed to change another user's info
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy2",
		},
		Password: &passwd,
	}))

	// And now try deleting the user
	require.Error(t, db.DelUser("testy2"))

	require.NoError(t, db.DelUser("testy"))
}

func TestAdminUser(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	// Create
	name := "testy"
	passwd := "testpass"
	require.NoError(t, adb.CreateUser(&User{
		UserName: &name,
		Password: &passwd,
	}))

	db := NewUserDB(adb, "testy")

	// Add testy to admin users
	adb.Assets().Config.AdminUsers = &[]string{"testy"}

	name2 := "testy2"
	require.NoError(t, db.CreateUser(&User{
		UserName: &name2,
		Password: &passwd,
	}))

	require.NoError(t, db.DelUser("testy2"))
}

func TestUserUpdateIcon(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	db := NewUserDB(adb, "testy")
	icon := "mi:lol"
	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID:   "testy",
			Icon: &icon,
		},
	}))
}

func TestUserObject(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	db := NewUserDB(adb, "testy")
	name := "tree"
	stype := "stream"
	sid, err := db.CreateObject(&Object{
		Details: Details{
			Name: &name,
		},
		Type: &stype,
	})
	require.NoError(t, err)

	name2 := "derpy"
	require.NoError(t, db.UpdateObject(&Object{
		Details: Details{
			ID:   sid,
			Name: &name2,
		},
		Meta: &JSONObject{
			"schema": 4,
		},
	}))
	s, err := db.ReadObject(sid, nil)
	require.NoError(t, err)
	require.Equal(t, *s.Name, name2)
	require.NotNil(t, s.OwnerScope)
	require.NotNil(t, s.Meta)
	require.True(t, s.Access.HasScope("*"))

	require.NoError(t, db.DelObject(sid))
	require.Error(t, db.DelObject(sid))
}
