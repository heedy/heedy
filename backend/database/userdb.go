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
	return createUser(db.adb, u, `SELECT 1 FROM scopesets WHERE scope='users:create' AND (
			name IN (
				SELECT scopeset FROM user_scopesets WHERE user=?
			) OR name IN ('public', 'users')
		) LIMIT 1;`, db.user)
}

func (db *UserDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A user can be read if:
	//	the user's public_access is >= 100 (read access by public),
	//	the user's user_access >=100
	//	the user is member of a group which gives it users:read scope
	//	the user to be read is itself, and the user has user:read scope
	if name != db.user {
		return readUser(db.adb, name, o, `SELECT * FROM users WHERE name=?
		AND (
				(public_access >= 100 OR user_access >=100)
			OR EXISTS 
				(SELECT 1 FROM scopesets WHERE scope='users:read' AND (
						name IN ('public', 'users') 
					OR 
						name IN (SELECT scopeset FROM user_scopesets WHERE user=?)
					)
				)
		)
		LIMIT 1;`, name, db.user)

	}

	return readUser(db.adb, name, o, `SELECT * FROM users WHERE name=?
		AND (
				(public_access >= 100 OR user_access >=100)
			OR EXISTS 
				(SELECT 1 FROM scopesets WHERE scope IN ('users:read', 'user:read') AND (
						name IN ('public', 'users') 
					OR 
						name IN (SELECT scopeset FROM user_scopesets WHERE user=?)
					)
				)
		)
		LIMIT 1;`, name, db.user)

}

// UpdateUser updates the given portions of a user
func (db *UserDB) UpdateUser(u *User) error {
	if u.ID == db.user {
		return updateUser(db.adb, u, `SELECT DISTINCT(scope) FROM scopesets WHERE (
					(scope LIKE 'users:edit%' OR scope LIKE 'user:edit%')
				AND
					(name IN (SELECT scopeset FROM user_scopesets WHERE user=?) OR name IN ('public', 'users'))
			);`, db.user)
	}
	return updateUser(db.adb, u, `SELECT DISTINCT(scope) FROM scopesets WHERE (
				scope LIKE 'users:edit%'
			AND
				(name IN (SELECT scopeset FROM user_scopesets WHERE user=?) OR name IN ('public', 'users'))
		);`, db.user)
}

func (db *UserDB) DelUser(name string) error {
	// A user can be deleted if:
	//	the user is member of a group which gives it users:delete scope
	//	the user to be read is itself, and the user has user:delete scope

	if db.user != name {
		return delUser(db.adb, name, `DELETE FROM users WHERE name=? AND EXISTS (
			SELECT 1 FROM scopesets WHERE scope='users:delete' AND (
				name IN ('public', 'users') 
				OR name IN (SELECT scopeset FROM user_scopesets WHERE user=?)
			) LIMIT 1
		);`, name, db.user)
	}

	// The user has same name, add check for user:delete
	return delUser(db.adb, name, `DELETE FROM users WHERE name=? AND EXISTS (
			SELECT 1 FROM scopesets WHERE scope IN ('user:delete', 'users:delete') AND (
				name IN ('public', 'users') 
				OR name IN (SELECT scopeset FROM user_scopesets WHERE user=?)
			) LIMIT 1	
		);`, name, db.user, db.user, db.user)
}

func (db *UserDB) ReadUserScopes(username string) ([]string, error) {
	if db.user != username {
		var scopes []string
		err := db.adb.Select(&scopes, `SELECT DISTINCT(scope) FROM scopesets WHERE
			(name IN ('public', 'users') OR name IN (SELECT scopeset FROM user_scopesets WHERE user=?))
			AND EXISTS (
				SELECT 1 FROM scopesets WHERE scope='users:scopes' AND (
					name IN ('public', 'users') 
					OR name IN (SELECT scopeset FROM user_scopesets WHERE user=?)
				)
			);`, username, db.user)
		if err == nil && len(scopes) == 0 {
			// TODO: This assumes that all users have at least one scope. Run another query here to make sure
			// that we don't have the necessary permissions.
			return scopes, ErrAccessDenied("You do not have permission to read scopes for '%s'", username)
		}
		return scopes, err
	}
	// If the user is me, just get my scopes, and check if I have the necessary permissions manually
	scopes, err := db.adb.ReadUserScopes(username)
	if err != nil {
		return scopes, err
	}
	for _, v := range scopes {
		if v == "users:scopes" || v == "user:scopes" {
			return scopes, nil
		}
	}

	return []string{}, ErrAccessDenied("You do not have permission to read your user's scopes")
}
