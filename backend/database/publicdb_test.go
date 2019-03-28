package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicUser(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	p := PublicDB{db}

	name := "testy"

	// Can't create the user
	require.EqualError(t, p.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}), ErrAccessDenied.Error())

	// Add user creation permission
	db.AddGroupScopes("public", "user:create")

	// Create
	require.NoError(t, p.CreateUser(&User{
		Group: Group{
			Details: Details{
				Name: &name,
			},
		},
		Password: "testpass",
	}))

}
