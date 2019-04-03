package database

type PublicDB struct {
	DB *AdminDB
}

func NewPublicDB(db *AdminDB) *PublicDB {
	return &PublicDB{DB: db}
}

// AdminDB returns the admin database
func (db *PublicDB) AdminDB() *AdminDB {
	return db.DB
}

func (db *PublicDB) ID() string {
	return ""
}

// User returns the user that is logged in
func (db *PublicDB) User() (*User, error) {
	return nil, nil
}

func (db *PublicDB) CreateUser(u *User) error {

	// Only create the user if the public group contains the user:create scope
	return createUser(db.DB, u, "SELECT 1 FROM group_scopes WHERE groupid='public' and scope='users:create';")

}

func (db *PublicDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A user can be read if the user's public_access is >= 100 (read access by public)
	// or if the public group has the users:read scope
	return readUser(db.DB, name, o, `SELECT * FROM groups WHERE id=? AND owner=id AND (
			public_access >= 100 
		OR 
			EXISTS (SELECT 1 FROM group_scopes WHERE groupid='public' and scope='users:read')
		) LIMIT 1;`, name)

}

func (db *PublicDB) UpdateUser(u *User) error {
	return updateUser(db.DB, u, `SELECT DISTINCT(scope) FROM group_scopes WHERE groupid='public' AND scope LIKE 'users:edit%';`)
}

func (db *PublicDB) DelUser(name string) error {
	return delUser(db.DB, name, `DELETE FROM users WHERE name=? AND EXISTS (
			SELECT 1 FROM group_scopes WHERE scope='users:delete' AND groupid='public'
		);`, name)
}

func (db *PublicDB) GetUserScopes(username string) ([]string, error) {

	var scopes []string
	err := db.DB.Select(&scopes, `SELECT DISTINCT(scope) FROM group_scopes WHERE
			(groupid IN (?, 'public', 'users') OR groupid IN (SELECT groupid FROM group_members WHERE username=?))
			AND EXISTS (SELECT 1 FROM group_scopes WHERE scope='users:scopes' AND groupid='public');`, username, username)
	if err == nil && len(scopes) == 0 {
		// TODO: Same error as in userdb - need to perform an additional check here, because maybe the user actually has 0 scopes (unlikely)
		return scopes, ErrAccessDenied
	}
	return scopes, err

}
