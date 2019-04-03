package database

import (
	"os"
	"testing"

	"github.com/heedy/heedy/backend/assets"

	"github.com/stretchr/testify/require"
)

func newAssets(t *testing.T) (*assets.Assets, func()) {
	a, err := assets.Open("", nil)
	require.NoError(t, err)
	os.RemoveAll("./test_db")
	a.FolderPath = "./"
	sqla := "sqlite3://test_db/heedy.db?_journal=WAL&_fk=1"
	a.Config.SQL = &sqla
	return a, func() {
		//os.RemoveAll("./test_db")
	}
}

func newDB(t *testing.T) (*AdminDB, func()) {
	a, cleanup := newAssets(t)

	err := Create(a)
	if err != nil {
		cleanup()
	}
	require.NoError(t, err)

	db, err := Open(a)
	require.NoError(t, err)

	return db, cleanup
}

func TestAdminUser(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	require.NoError(t, db.CreateUser(&User{

		Details: Details{
			Name: &name,
		},

		Password: &passwd,
	}))

	name = "test2"
	require.EqualError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
	}), ErrNoPasswordGiven.Error())

	name = "tee hee"
	passwd = "mypass"
	require.Error(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}), "Bad name must fail validation")

	name = "testy"
	require.Error(t, db.CreateUser(&User{

		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}), "Should fail to add existing user")

	_, err := db.ReadUser("testee", nil)
	require.EqualError(t, err, ErrUserNotFound.Error())

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")
	require.Nil(t, u.Password, "The password should never be read back")

	_, _, err = db.AuthUser("testy", "testpass")
	require.NoError(t, err)

	_, _, err = db.AuthUser("testy", "testpass2")
	require.Error(t, err)

	_, _, err = db.AuthUser("testerr", "mypass")
	require.Error(t, err)

	name = "testy"
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: name,
		},
	}), "Updating nothing should give an error")

	passwd = "mypass2"
	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: name,
		},
		Password: &passwd,
	}), "Update password should succeed")

	_, _, err = db.AuthUser("testy", "testpass")
	require.Error(t, err, "Password should have been changed")

	_, _, err = db.AuthUser("testy", "mypass2")
	require.NoError(t, err, "This should be the new password...")

	name2 := "testyeee"
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: name2,
		},
		Password: &passwd,
	}), "Update should fail on nonexistent user")

	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID:   name,
			Name: &name2,
		},
	}), "User name should update")

	_, err = db.ReadUser(name, nil)
	require.Error(t, err)
	u, err = db.ReadUser(name2, nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, name2)

	require.NoError(t, db.DelUser(name2))

	require.Error(t, db.DelUser(name2), "Deleting nonexistent user should fail")

	_, _, err = db.AuthUser("testy", "mypass2")
	require.Error(t, err, "User should no longer exist")

}

func TestAdminGroup(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	_, err := db.ReadGroup("testy", nil)
	require.Error(t, err, "A user is not a group")

	gdesc := "This is a testy group"
	gid, err := db.CreateGroup(&Group{
		Details: Details{
			Name:        &name,
			Description: &gdesc,
		},
		Owner: &name,
	})
	require.NoError(t, err)

	g, err := db.ReadGroup(gid, nil)
	require.NoError(t, err, "A group should be selectable")
	require.NotNil(t, g.Description)
	require.Equal(t, gdesc, *g.Description)

	_, err = db.ReadGroup("tree", nil)
	require.Error(t, err, "Group should not exist")

	owner := "derp"
	err = db.UpdateGroup(&Group{
		Details: Details{
			ID: gid,
		},
		Owner: &owner,
	})
	require.Error(t, err, "Group owner must be valid")

	err = db.DelGroup("testy")
	require.Error(t, err, "Deleting user must fail")

	err = db.DelGroup(gid)
	require.NoError(t, err)

	_, err = db.ReadGroup(gid, nil)
	require.Error(t, err, "Group should not exist")

}

func TestAdminConnection(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	badname := "derp"
	conn, apikey, err := db.CreateConnection(&Connection{
		Details: Details{
			Name: &name,
		},
		Owner: &name,
	})
	require.NoError(t, err)
	require.Equal(t, apikey, "")

	_, err = db.GetConnectionByKey("")
	require.Error(t, err)

	c, err := db.ReadConnection(conn, nil)
	require.NoError(t, err)

	require.Equal(t, *c.Name, name)

	c = &Connection{
		Details: Details{
			ID: conn,
		},
		APIKey: &badname, // can be anything
	}
	require.NoError(t, db.UpdateConnection(c))
	require.NotEqual(t, badname, *c.APIKey, "The API key should have been changed during update")

	c2, err := db.GetConnectionByKey(*c.APIKey)
	require.NoError(t, err)
	require.Equal(t, c2.ID, c.ID)

	require.NoError(t, db.DelConnection(c.ID))

	_, err = db.ReadConnection(conn, nil)
	require.Error(t, err)
}

func TestAdminStream(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	badname := "derp"
	conn, _, err := db.CreateConnection(&Connection{
		Details: Details{
			Name: &name,
		},
		Owner: &name,
	})
	require.NoError(t, err)
	sid, err := db.CreateStream(&Stream{
		Details: Details{

			Name: &name,
		},
		User:       &name,
		Connection: &conn,
	})
	require.NoError(t, err)

	require.NoError(t, db.UpdateStream(&Stream{
		Details: Details{
			ID:       sid,
			FullName: &badname,
		},
	}))

	s, err := db.ReadStream(sid, nil)
	require.NoError(t, err)
	require.Equal(t, *s.FullName, badname)

	require.NoError(t, db.DelStream(sid))
	require.Error(t, db.DelStream(sid))
}

func TestAdminScopes(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: &passwd,
	}))

	scopesets, err := db.ReadUserScopeSets("testy")
	require.NoError(t, err)
	require.Contains(t, scopesets, "users")
	require.Contains(t, scopesets, "public")

	require.NoError(t, db.AddUserScopeSets("testy", "tee", "hee", "public")) // public scope set won't be added
	scopesets2, err := db.ReadUserScopeSets("testy")
	require.NoError(t, err)
	require.Equal(t, len(scopesets), len(scopesets2)-2)

	err = db.RemUserScopeSets("testy", "users")
	require.Error(t, err)
	err = db.RemUserScopeSets("testy", "hee")
	require.NoError(t, err)
	scopesets2, err = db.ReadUserScopeSets("testy")
	require.NoError(t, err)
	require.Equal(t, len(scopesets), len(scopesets2)-1)

	scopes, err := db.ReadUserScopes("testy")
	require.NoError(t, err)
	require.True(t, len(scopes) > 0)

	s, err := db.ReadScopeSet("tee")
	require.NoError(t, err)
	require.Equal(t, len(s), 0)

	err = db.AddScopeSet("tee", "hee")
	require.NoError(t, err)
	s, err = db.ReadScopeSet("tee")
	require.NoError(t, err)
	require.Equal(t, len(s), 1)
	require.Equal(t, s[0], "hee")

	scopes2, err := db.ReadUserScopes("testy")
	require.NoError(t, err)
	require.Equal(t, len(scopes), len(scopes2)-1)

	require.NoError(t, db.AddUserScopeSets("testy", "scree"))
	ss, err := db.GetAllScopeSets()
	require.NoError(t, err)

	require.Equal(t, len(ss), 4) // users,public, tee, scree

	require.NoError(t, db.DeleteScopeSet("tee"))

	scopesets, err = db.ReadUserScopeSets("testy")
	require.NoError(t, err)
	require.NotContains(t, scopesets, "tee")

	ss, err = db.GetAllScopeSets()
	require.NoError(t, err)
	require.NotContains(t, ss, "tee")

}
