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
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

	name = "test2"
	require.EqualError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "",
	}), ErrNoPasswordGiven.Error())

	name = "tee hee"
	require.Error(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "mypass",
	}), "Bad name must fail validation")

	name = "testy"
	require.Error(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "mypass",
	}), "Should fail to add existing user")

	_, err := db.ReadUser("testee")
	require.EqualError(t, err, ErrUserNotFound.Error())

	u, err := db.ReadUser("testy")
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	_, _, err = db.AuthUser("testy", "testpass")
	require.NoError(t, err)

	_, _, err = db.AuthUser("testy", "testpass2")
	require.Error(t, err)

	_, _, err = db.AuthUser("testerr", "mypass")
	require.Error(t, err)

	name = "testy"
	require.Error(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: name,
			},
		},
	}), "Updating nothing should give an error")

	require.NoError(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: name,
			},
		},
		Password: "mypass2",
	}), "Update password should succeed")

	_, _, err = db.AuthUser("testy", "testpass")
	require.Error(t, err, "Password should have been changed")

	_, _, err = db.AuthUser("testy", "mypass2")
	require.NoError(t, err, "This should be the new password...")

	name2 := "testyeee"
	require.Error(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID: name2,
			},
		},
		Password: "mypass2",
	}), "Update should fail on nonexistent user")

	require.NoError(t, db.UpdateUser(&User{
		Group: Group{
			Details: Details{
				ID:   name,
				Name: &name2,
			},
		},
	}), "User name should update")

	_, err = db.ReadUser(name)
	require.Error(t, err)
	u, err = db.ReadUser(name2)
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
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

	ug, err := db.ReadGroup("testy")
	require.NoError(t, err, "A user is a group")
	require.Equal(t, *ug.Name, name)

	gdesc := "This is a testy group"
	gid, err := db.CreateGroup(&Group{
		Details: Details{
			Name:        &name,
			Description: &gdesc,
			Owner:       &name,
		},
	})
	require.NoError(t, err)

	g, err := db.ReadGroup(gid)
	require.NoError(t, err, "A group should be selectable")
	require.NotNil(t, g.Description)
	require.Equal(t, gdesc, *g.Description)

	_, err = db.ReadGroup("tree")
	require.Error(t, err, "Group should not exist")

	owner := "derp"
	err = db.UpdateGroup(&Group{
		Details: Details{
			ID:    gid,
			Owner: &owner,
		},
	})
	require.Error(t, err, "Group owner must be valid")

	err = db.DelGroup("testy")
	require.Error(t, err, "Deleting user must fail")

	err = db.DelGroup(gid)
	require.NoError(t, err)

	_, err = db.ReadGroup(gid)
	require.Error(t, err, "Group should not exist")

}

func TestAdminConnection(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

	badname := "derp"
	conn, apikey, err := db.CreateConnection(&Connection{
		Details: Details{
			Name:  &name,
			Owner: &name,
		},
	})
	require.NoError(t, err)
	require.Equal(t, apikey, "")

	_, err = db.GetConnectionByKey("")
	require.Error(t, err)

	c, err := db.ReadConnection(conn)
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

	_, err = db.ReadConnection(conn)
	require.Error(t, err)
}

func TestAdminStream(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

	badname := "derp"
	conn, _, err := db.CreateConnection(&Connection{
		Details: Details{
			Name:  &name,
			Owner: &name,
		},
	})
	require.NoError(t, err)
	sid, err := db.CreateStream(&Stream{
		Details: Details{
			Owner: &name,
			Name:  &name,
		},
		Connection: &conn,
	})
	require.NoError(t, err)

	require.NoError(t, db.UpdateStream(&Stream{
		Details: Details{
			ID:       sid,
			FullName: &badname,
		},
	}))

	s, err := db.ReadStream(sid)
	require.NoError(t, err)
	require.Equal(t, *s.FullName, badname)

	require.NoError(t, db.DelStream(sid))
	require.Error(t, db.DelStream(sid))
}

func TestAdminScopes(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

	scopes, err := db.GetGroupScopes(name)
	require.NoError(t, err)
	require.Equal(t, 0, len(scopes))

	scopes, err = db.GetGroupScopes("notvalid")
	require.NoError(t, err)
	require.Equal(t, 0, len(scopes))

	require.Error(t, db.AddGroupScopes("notvalid", "myscope", "myscope2"))

	scopes, err = db.GetGroupScopes("notvalid")
	require.NoError(t, err)
	require.Equal(t, 0, len(scopes))

	require.NoError(t, db.AddGroupScopes(name, "myscope", "myscope2"))

	scopes, err = db.GetGroupScopes(name)
	require.NoError(t, err)
	require.Equal(t, 2, len(scopes))

	require.NoError(t, db.RemGroupScopes("notvalid", "myscope"))

	require.NoError(t, db.RemGroupScopes(name, "myscope3", "myscope"))

	scopes, err = db.GetGroupScopes(name)
	require.NoError(t, err)
	require.Equal(t, 1, len(scopes))
	require.Equal(t, "myscope2", scopes[0])
}
