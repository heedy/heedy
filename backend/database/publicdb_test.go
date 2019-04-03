package database

/*
func TestPublicUser(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	db := PublicDB{adb}

	name := "testy"

	// Can't create the user
	require.EqualError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: "testpass",
	}), ErrAccessDenied.Error())

	// Add user creation permission
	adb.AddGroupScopes("public", "users:create")

	// Create
	require.NoError(t, db.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: "testpass",
	}))
	_, err := db.ReadUser("testy", nil)
	require.Error(t, err)

	adb.AddGroupScopes("public", "users:read")

	u, err := db.ReadUser("testy", nil)
	require.NoError(t, err)
	require.Equal(t, *u.Name, "testy")

	// Shouldn't be allowed to change another user's password without the scope present
	require.Error(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: "mypass2",
	}))

	adb.AddGroupScopes("public", "users:edit:password")

	require.NoError(t, db.UpdateUser(&User{
		Details: Details{
			ID: "testy",
		},
		Password: "mypass2",
	}))

	require.Error(t, db.DelUser("testy"))
	adb.AddGroupScopes("public", "users:delete")
	require.NoError(t, db.DelUser("testy"))

	_, err = adb.ReadUser("testy", nil)
	require.Error(t, err)
}

func TestPublicUserScope(t *testing.T) {
	adb, cleanup := newDB(t)
	defer cleanup()

	// Create
	name := "testy"
	require.NoError(t, adb.CreateUser(&User{
		Details: Details{
			Name: &name,
		},
		Password: "testpass",
	}))

	db := NewPublicDB(adb)

	_, err := db.GetUserScopes("testy")
	require.Error(t, err)

	require.NoError(t, adb.AddGroupScopes("public", "users:scopes"))

	_, err = db.GetUserScopes("testy")
	require.NoError(t, err)
}
*/
