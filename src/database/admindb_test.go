package database

import (
	"os"
	"testing"

	"github.com/connectordb/connectordb/assets"

	"github.com/stretchr/testify/require"
)

func newAssets(t *testing.T) (*assets.Assets, func()) {
	a, err := assets.Open("", nil)
	require.NoError(t, err)
	a.FolderPath = "./"
	sqla := "sqlite3://test_db/cdb.db?_journal=WAL"
	a.Config.SQL = &sqla
	return a, func() {
		os.RemoveAll("./test_db")
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
			Name: &name,
		},
		Password: "testpass",
	}))

	name = "test2"
	require.EqualError(t, db.CreateUser(&User{
		Group: Group{
			Name: &name,
		},
		Password: "",
	}), ErrNoPasswordGiven.Error())

	name = "tee hee"
	require.Error(t, db.CreateUser(&User{
		Group: Group{
			Name: &name,
		},
		Password: "mypass",
	}), "Bad name must fail validation")

	name = "testy"
	require.Error(t, db.CreateUser(&User{
		Group: Group{
			Name: &name,
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
			ID: name,
		},
	}), "Updating nothing should give an error")

	require.NoError(t, db.UpdateUser(&User{
		Group: Group{
			ID: name,
		},
		Password: "mypass2",
	}), "Update password should succeed")

	_, _, err = db.AuthUser("testy", "testpass")
	require.Error(t, err, "Password should have been changed")

	_, _, err = db.AuthUser("testy", "mypass2")
	require.NoError(t, err, "This should be the new password...")

	name = "testyeee"
	require.Error(t, db.UpdateUser(&User{
		Group: Group{
			Name: &name,
		},
		Password: "mypass2",
	}), "Update should fail on nonexistent user")

	require.NoError(t, db.DelUser("testy"))

	require.Error(t, db.DelUser("testy"), "Deleting nonexistent user should fail")

	_, _, err = db.AuthUser("testy", "mypass2")
	require.Error(t, err, "User should no longer exist")

}

func TestAdminGroup(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	name := "testy"
	require.NoError(t, db.CreateUser(&User{
		Group: Group{
			Name: &name,
		},
		Password: "testpass",
	}))

	ug, err := db.ReadGroup("testy")
	require.NoError(t, err, "A user is a group")
	require.Equal(t, *ug.Name, name)

	gdesc := "This is a testy group"
	gid, err := db.CreateGroup(&Group{
		Name:        &name,
		Description: &gdesc,
		Owner:       &name,
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
		ID:    gid,
		Owner: &owner,
	})
	require.Error(t, err, "Group owner must be valid")

	err = db.DelGroup("testy")
	require.Error(t, err, "Deleting user must fail")

	err = db.DelGroup(gid)
	require.NoError(t, err)

	_, err = db.ReadGroup(gid)
	require.Error(t, err, "Group should not exist")

}

func TestAdminGroupPermissions(t *testing.T) {

}
