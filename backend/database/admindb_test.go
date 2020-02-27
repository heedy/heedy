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
			ID: name,
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

func TestAdminApp(t *testing.T) {
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
	conn, AccessToken, err := db.CreateApp(&App{
		Details: Details{
			Name: &name,
		},
		Owner:       &name,
		AccessToken: &noAccessToken,
	})
	require.NoError(t, err)
	require.Equal(t, AccessToken, "")

	_, err = db.GetAppByAccessToken("")
	require.Error(t, err)

	c, err := db.ReadApp(conn, nil)
	require.NoError(t, err)

	require.Equal(t, *c.Name, name)

	c = &App{
		Details: Details{
			ID: conn,
		},
		AccessToken: &badname, // can be anything
	}
	require.NoError(t, db.UpdateApp(c))
	require.NotEqual(t, badname, *c.AccessToken, "The API key should have been changed during update")

	c2, err := db.GetAppByAccessToken(*c.AccessToken)
	require.NoError(t, err)
	require.Equal(t, c2.ID, c.ID)

	require.NoError(t, db.DelApp(c.ID))

	_, err = db.ReadApp(conn, nil)
	require.Error(t, err)
}

func TestAdminObject(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	passwd := "testpass"
	stype := "timeseries"
	require.NoError(t, db.CreateUser(&User{
		UserName: &name,
		Password: &passwd,
	}))

	badname := "derp"
	conn, _, err := db.CreateApp(&App{
		Details: Details{
			Name: &name,
		},
		Owner: &name,
	})
	require.NoError(t, err)
	sid, err := db.CreateObject(&Object{
		Details: Details{

			Name: &name,
		},
		App:  &conn,
		Type: &stype,
	})
	require.NoError(t, err)

	require.NoError(t, db.UpdateObject(&Object{
		Details: Details{
			ID:   sid,
			Name: &badname,
		},

		Meta: &JSONObject{
			"schema": 4,
		},

		OwnerScope: &ScopeArray{
			Scope: []string{"myscope1", "myscope2"},
		},
	}))

	s, err := db.ReadObject(sid, nil)
	require.NoError(t, err)
	require.Equal(t, *s.Name, badname)
	require.NotNil(t, s.OwnerScope)
	require.NotNil(t, s.Meta)
	require.Equal(t, len((*s.OwnerScope).Scope), 2)
	require.Equal(t, (*s.Meta)["schema"], float64(4))

	sl, err := db.ListObjects(nil)
	require.NoError(t, err)
	require.Equal(t, len(sl), 1)
	objectType := "notascouce"
	sl, err = db.ListObjects(&ListObjectsOptions{
		Type: &objectType,
	})
	require.NoError(t, err)
	require.Equal(t, len(sl), 0)
	sl, err = db.ListObjects(&ListObjectsOptions{
		Type: &stype,
	})
	require.NoError(t, err)
	require.Equal(t, len(sl), 1)
	//fmt.Printf(s.String())

	require.NoError(t, db.DelObject(sid))
	require.Error(t, db.DelObject(sid))
}

func TestAdminShareObject(t *testing.T) {
	db, cleanup := newDBWithUser(t)
	defer cleanup()

	name := "testy"
	stype := "timeseries"
	sid, err := db.CreateObject(&Object{
		Details: Details{

			Name: &name,
		},
		Owner: &name,
		Type:  &stype,
	})
	require.NoError(t, err)

	m, err := db.GetObjectShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 0)

	require.NoError(t, db.ShareObject(sid, "public", &ScopeArray{
		Scope: []string{"read", "write"},
	}))

	m, err = db.GetObjectShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 1)

	require.Equal(t, len(m["public"].Scope), 2)

	require.NoError(t, db.ShareObject(sid, "users", &ScopeArray{
		Scope: []string{"read", "write", "love"},
	}))

	m, err = db.GetObjectShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 2)

	require.NoError(t, db.UnshareObjectFromUser(sid, "users"))

	m, err = db.GetObjectShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 1)

	require.NoError(t, db.UnshareObject(sid))

	m, err = db.GetObjectShares(sid)
	require.NoError(t, err)
	require.Equal(t, len(m), 0)
}

func TestKey(t *testing.T) {
	db, cleanup := newDBWithUser(t)
	defer cleanup()

	name := "testy"
	otype := "timeseries"

	// Key can't be set for non-app objects
	_, err := db.CreateObject(&Object{
		Details: Details{

			Name: &name,
		},
		Key:   &name,
		Owner: &name,
		Type:  &otype,
	})
	require.Error(t, err)

	appid, _, err := db.CreateApp(&App{
		Details: Details{
			Name: &name,
		},
		Owner: &name,
	})
	require.NoError(t, err)

	oid1, err := db.CreateObject(&Object{
		Details: Details{

			Name: &name,
		},
		App:  &appid,
		Key:  &name,
		Type: &otype,
	})
	require.NoError(t, err)

	// App keys are unique
	_, err = db.CreateObject(&Object{
		Details: Details{

			Name: &name,
		},
		App:  &appid,
		Key:  &name,
		Type: &otype,
	})
	require.Error(t, err)

	// Allow creating different key
	key2 := "key2"
	oid2, err := db.CreateObject(&Object{
		Details: Details{

			Name: &name,
		},
		App:  &appid,
		Key:  &key2,
		Type: &otype,
	})
	require.NoError(t, err)

	// Allow querying by key
	objs, err := db.ListObjects(&ListObjectsOptions{
		Key: &key2,
	})

	require.NoError(t, err)
	require.Len(t, objs, 1)
	require.Equal(t, objs[0].ID, oid2)
	require.Equal(t, *objs[0].Key, key2)

	// Now remove the key from oid1, and query objects with no key
	es := ""
	err = db.UpdateObject(&Object{
		Details: Details{
			ID: oid1,
		},
		Key: &es,
	})
	require.NoError(t, err)

	objs, err = db.ListObjects(&ListObjectsOptions{
		Key: &es,
	})

	require.NoError(t, err)
	require.Len(t, objs, 1)
	require.Equal(t, objs[0].ID, oid1)
	require.Nil(t, objs[0].Key)

}

func TestTags(t *testing.T) {
	db, cleanup := newDBWithUser(t)
	defer cleanup()

	name := "testy"
	otype := "timeseries"

	tags := &StringArray{Strings: []string{"tag1", "tag2", "tag3"}}
	// Key can't be set for non-app objects
	oid1, err := db.CreateObject(&Object{
		Details: Details{

			Name: &name,
		},
		Tags:  tags,
		Owner: &name,
		Type:  &otype,
	})
	require.NoError(t, err)

	stags := tags.String()
	objs, err := db.ListObjects(&ListObjectsOptions{
		Tags: &stags,
	})

	require.NoError(t, err)
	require.Len(t, objs, 1)
	require.Equal(t, objs[0].ID, oid1)

	stags = "tag1 tag4"

	objs, err = db.ListObjects(&ListObjectsOptions{
		Tags: &stags,
	})

	require.NoError(t, err)
	require.Len(t, objs, 0)

	stags = "tag3 tag1"

	objs, err = db.ListObjects(&ListObjectsOptions{
		Tags: &stags,
	})

	require.NoError(t, err)
	require.Len(t, objs, 1)
	require.Equal(t, objs[0].ID, oid1)
}
