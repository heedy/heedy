package database

import (
	"testing"

	"github.com/heedy/heedy/backend/database/dbutil"
	"github.com/stretchr/testify/require"
)

func TestAppObject(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	udb := NewUserDB(adb, "testy")

	cname := "conn"
	cid, _, err := udb.CreateApp(&App{
		Details: Details{
			ID:   "conn",
			Name: &cname,
		},
		Scope: &AppScopeArray{
			ScopeArray: ScopeArray{
				Scope: []string{"self.objects.timeseries", "owner:read"},
			},
		},
	})
	require.NoError(t, err)
	c, err := udb.ReadApp(cid, nil)
	require.NoError(t, err)
	cdb := NewAppDB(adb, c)

	name := "tree"
	stype := "timeseries"
	sid, err := cdb.CreateObject(&Object{
		Details: Details{
			Name: &name,
		},
		Type: &stype,
	})
	require.NoError(t, err)

	name2 := "derpy"
	require.Error(t, cdb.UpdateObject(&Object{
		Details: Details{
			ID:   sid,
			Name: &name2,
		},
		Meta: &dbutil.JSONObject{
			"schema": 4,
		},
	}))
	require.NoError(t, cdb.UpdateObject(&Object{
		Details: Details{
			ID:   sid,
			Name: &name2,
		},
		Meta: &dbutil.JSONObject{
			"actor": true,
		},
	}))

	s, err := cdb.ReadObject(sid, nil)
	require.NoError(t, err)
	require.Equal(t, *s.Name, name2)
	require.NotNil(t, s.OwnerScope)
	require.NotNil(t, s.Meta)
	require.True(t, s.Access.HasScope("*"))

	require.NoError(t, cdb.DelObject(sid))
	require.Error(t, cdb.DelObject(sid))
}
