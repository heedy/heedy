package database

type UserDB struct {
	adb *AdminDB

	user string
}

func NewUserDB(adb *AdminDB, user string) *UserDB {
	return &UserDB{
		adb:  adb,
		user: user,
	}
}

// AdminDB returns the admin database
func (db *UserDB) AdminDB() *AdminDB {
	return db.adb
}

func (db *UserDB) ID() string {
	return db.user
}

// User returns the user that is logged in
func (db *UserDB) User() (*User, error) {
	return db.ReadUser(db.user, &ReadUserOptions{
		Avatar: true,
	})
}

func (db *UserDB) CreateUser(u *User) error {
	// Only create the user if I have the users:create scope
	return createUser(db.adb, u, `SELECT 1 FROM user_scopes WHERE user=? AND scope='users:create';`, db.user)
}

func (db *UserDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A user can be read if it is the current user, OR
	//	the user's public_access is >= 100 (read access by public),
	//	the user's user_access >=100
	//	current user has users:read scope
	if name == db.user {
		return db.adb.ReadUser(name, o)
	}

	return readUser(db.adb, name, o, `SELECT * FROM users WHERE name=? AND EXISTS 
		(SELECT 1 FROM user_can_read_user WHERE target=? AND user=?) LIMIT 1;`, name, name, db.user)

}

// UpdateUser updates the given portions of a user
func (db *UserDB) UpdateUser(u *User) error {
	if u.ID == db.user {
		if u.Name != nil {
			// Updating the username requires user:edit:name scope
			return updateUser(db.adb, u, `SELECT 1 FROM user_scopes 
					WHERE user=? AND scope IN ('user:edit:name', 'users:edit:name');`, db.user)
		}
		return db.adb.UpdateUser(u)
	}

	requiredScopes := []string{"users:edit"}

	if u.Password != nil {
		// Trying to edit the user's password
		requiredScopes = append(requiredScopes, "users:edit:password")
	}
	if u.Name != nil {
		requiredScopes = append(requiredScopes, "users:edit:name")
	}

	return updateUser(db.adb, u, sqlIn(`SELECT 1 FROM user_can_read_user WHERE user=? AND target=? 
			AND ?=(SELECT COUNT(DISTINCT(scope)) FROM user_scopes WHERE scope IN (%s));`, requiredScopes), db.user, u.ID, len(requiredScopes))
}

func (db *UserDB) DelUser(name string) error {
	// A user can be deleted if:
	//	the user is member of a group which gives it users:delete scope
	//	the user to be read is itself, and the user has user:delete scope
	if name == db.user {
		return delUser(db.adb, name, `DELETE FROM users WHERE name=? AND EXISTS (SELECT 1 FROM user_scopes WHERE user=? AND scope IN ('user:delete', 'users:delete'));`, name, name)
	}

	return delUser(db.adb, name, `DELETE FROM users WHERE name=? 
			AND EXISTS (SELECT 1 FROM user_can_read_user WHERE target=? AND user=?) 
			AND 'users:delete' IN (SELECT scope FROM user_scopes WHERE user=?)`, name, name, db.user, db.user)
}

func (db *UserDB) ReadUserScopes(username string) ([]string, error) {
	if db.user == username {
		return db.adb.ReadUserScopes(username)
	}
	return readUserScopes(db.adb, username, `SELECT 1 FROM user_can_read_user WHERE target=? AND user=?
		AND 'users:scopes' IN (SELECT scope FROM user_scopes s WHERE s.user=?);`, username, db.user, db.user)
}
