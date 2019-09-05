package database


import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionSource(t *testing.T) {
	adb, cleanup := newDBWithUser(t)
	defer cleanup()

	udb := NewUserDB(adb, "testy")

	cname := "conn"
	cid,_, err := udb.CreateConnection(&Connection{
		Details: Details{
			ID: "conn",
			Name: &cname,
		},
		Scopes: &ConnectionScopeArray{
			ScopeArray: ScopeArray{
				Scopes: []string{"self.sources.stream","owner:read"},
			},
		},
	})
	require.NoError(t,err)
	c,err := udb.ReadConnection(cid,nil)
	require.NoError(t,err)
	cdb := NewConnectionDB(adb,c)

	name := "tree"
	stype := "stream"
	sid, err := cdb.CreateSource(&Source{
		Details: Details{
			Name: &name,
		},
		Type: &stype,
	})
	require.NoError(t, err)

	name2 := "derpy"
	require.NoError(t, cdb.UpdateSource(&Source{
		Details: Details{
			ID:       sid,
			Name: &name2,
		},
		Meta: &SourceMeta{
			"schema": 4,
		},
	}))

	s, err := cdb.ReadSource(sid, nil)
	require.NoError(t, err)
	require.Equal(t, *s.Name, name2)
	require.NotNil(t, s.Scopes)
	require.NotNil(t, s.Meta)
	require.True(t, s.Access.HasScope("*"))

	require.NoError(t, cdb.DelSource(sid))
	require.Error(t, cdb.DelSource(sid))
}