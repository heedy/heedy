package group

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
