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
	a.FolderPath = "./test_db"
	sqla := "sqlite3://heedy.db?_journal=WAL&_fk=1"
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

func newDBWithUser(t *testing.T) (*AdminDB, func()) {
	adb, cleanup := newDB(t)

	name := "testy"
	passwd := "testpass"
	require.NoError(t, adb.CreateUser(&User{
		UserName: &name,
		Password: &passwd,
	}))
	return adb, cleanup
}

func TestAdminDBUser(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	require.NoError(t, db.CreateUser(&User{
		UserName: &name,
		Password: &passwd,
	}))

	name = "test2"
	require.EqualError(t, db.CreateUser(&User{
		UserName: &name,
	}), ErrNoPasswordGiven.Error())

	name = "tee hee"
	passwd = "mypass"
	require.Error(t, db.CreateUser(&User{
		UserName: &name,
		Password: &passwd,
	}), "Bad name must fail validation")

	name = "testy"
	require.Error(t, db.CreateUser(&User{

		UserName: &name,
		Password: &passwd,
	}), "Should fail to add existing user")

	_, err := db.ReadUser("testee", nil)
	require.EqualError(t, err, ErrUserNotFound.Error())

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.UserName, "testy")
	// users don't have access
	//require.Equal(t, u.Access.String(), "read update update:password delete")
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
		},
		UserName: &name2,
	}), "User name should update")

	_, err = db.ReadUser(name, nil)
	require.Error(t, err)
	u, err = db.ReadUser(name2, nil)
	require.NoError(t, err)
	require.Equal(t, *u.UserName, name2)

	require.NoError(t, db.DelUser(name2))

	require.Error(t, db.DelUser(name2), "Deleting nonexistent user should fail")

	_, _, err = db.AuthUser("testy", "mypass2")
	require.Error(t, err, "User should no longer exist")

}

func TestAdminConnection(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	require.NoError(t, db.CreateUser(&User{
		UserName: &name,
		Password: &passwd,
	}))

	badname := "derp"
	noAccessToken := ""
	conn, AccessToken, err := db.CreateConnection(&Connection{
		Details: Details{
			Name: &name,
		},
		Owner: &name,
		AccessToken: &noAccessToken,
	})
	require.NoError(t, err)
	require.Equal(t, AccessToken, "")

	_, err = db.GetConnectionByAccessToken("")
	require.Error(t, err)

	c, err := db.ReadConnection(conn, nil)
	require.NoError(t, err)

	require.Equal(t, *c.Name, name)

	c = &Connection{
		Details: Details{
			ID: conn,
		},
		AccessToken: &badname, // can be anything
	}
	require.NoError(t, db.UpdateConnection(c))
	require.NotEqual(t, badname, *c.AccessToken, "The API key should have been changed during update")

	c2, err := db.GetConnectionByAccessToken(*c.AccessToken)
	require.NoError(t, err)
	require.Equal(t, c2.ID, c.ID)

	require.NoError(t, db.DelConnection(c.ID))

	_, err = db.ReadConnection(conn, nil)
	require.Error(t, err)
}

func TestAdminSource(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	stype := "stream"
	require.NoError(t, db.CreateUser(&User{
		UserName: &name,
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
	sid, err := db.CreateSource(&Source{
		Details: Details{

			Name: &name,
		},
		Connection: &conn,
		Type:       &stype,
	})
	require.NoError(t, err)

	require.NoError(t, db.UpdateSource(&Source{
		Details: Details{
			ID:       sid,
			Name: &badname,
		},

		Meta: &JSONObject{
			"schema": 4,
		},

		Scopes: &ScopeArray{
			Scopes: []string{"myscope1", "myscope2"},
		},
	}))

	s, err := db.ReadSource(sid, nil)
	require.NoError(t, err)
	require.Equal(t, *s.Name, badname)
	require.NotNil(t, s.Scopes)
	require.NotNil(t, s.Meta)
	require.Equal(t, len((*s.Scopes).Scopes), 2)
	require.Equal(t, (*s.Meta)["schema"], float64(4))

	sl,err := db.ListSources(nil)
	require.NoError(t,err)
	require.Equal(t,len(sl),1)
	sourceType:= "notascouce"
	sl,err = db.ListSources(&ListSourcesOptions{
		Type: &sourceType,
	})
	require.NoError(t,err)
	require.Equal(t,len(sl),0)
	sl,err = db.ListSources(&ListSourcesOptions{
		Type: &stype,
	})
	require.NoError(t,err)
	require.Equal(t,len(sl),1)
	//fmt.Printf(s.String())

	require.NoError(t, db.DelSource(sid))
	require.Error(t, db.DelSource(sid))
}

func TestAdminShareSource(t *testing.T) {
	db, cleanup := newDBWithUser(t)
	defer cleanup()

	name := "testy"
	stype := "stream"
	sid, err := db.CreateSource(&Source{
		Details: Details{

			Name: &name,
		},
		Owner: &name,
		Type:  &stype,
	})
	require.NoError(t, err)

	m, err := db.GetSourceShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 0)

	require.NoError(t, db.ShareSource(sid, "public", &ScopeArray{
		Scopes: []string{"read", "write"},
	}))

	m, err = db.GetSourceShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 1)

	require.Equal(t, len(m["public"].Scopes), 2)

	require.NoError(t, db.ShareSource(sid, "users", &ScopeArray{
		Scopes: []string{"read", "write", "love"},
	}))

	m, err = db.GetSourceShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 2)

	require.NoError(t, db.UnshareSourceFromUser(sid, "users"))

	m, err = db.GetSourceShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 1)

	require.NoError(t, db.UnshareSource(sid))

	m, err = db.GetSourceShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 0)
}
