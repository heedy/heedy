package database

type PublicDB struct {
	adb *AdminDB
}

func NewPublicDB(db *AdminDB) *PublicDB {
	return &PublicDB{adb: db}
}

// AdminDB returns the admin database
func (db *PublicDB) AdminDB() *AdminDB {
	return db.adb
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
	return createUser(db.adb, u, "SELECT 1 FROM scopesets WHERE name='public' and scope='users:create';")

}

func (db *PublicDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A user can be read if the user's public_access is >= 100 (read access by public)
	// or if the public group has the users:read scope
	return readUser(db.adb, name, o, `SELECT * FROM users WHERE name=? 
		AND EXISTS (SELECT 1 FROM public_can_read_user WHERE user=?);`, name, name)

}

func (db *PublicDB) UpdateUser(u *User) error {

	requiredScopes := []string{"users:edit"}

	if u.Password != nil {
		requiredScopes = append(requiredScopes, "users:edit:password")
	}
	if u.Name != nil {
		requiredScopes = append(requiredScopes, "users:edit:name")
	}

	return updateUser(db.adb, u, sqlIn(`SELECT 1 FROM public_can_read_user WHERE user=?
			AND ?=(SELECT COUNT(DISTINCT(scope)) FROM scopesets WHERE name='public' AND scope IN (%s));`, requiredScopes), u.ID, len(requiredScopes))
}

func (db *PublicDB) DelUser(name string) error {
	return delUser(db.adb, name, `DELETE FROM users WHERE name=?
			AND EXISTS (SELECT 1 FROM public_can_read_user WHERE user=?)
			AND EXISTS (SELECT 1 FROM scopesets WHERE scope='users:delete' AND name='public');`, name, name)
}

func (db *PublicDB) ReadUserScopes(username string) ([]string, error) {
	return readUserScopes(db.adb, username, `SELECT 1 FROM public_can_read_user WHERE user=?
			AND EXISTS (SELECT 1 FROM scopesets WHERE scope='users:scopes' AND name='public');`, username, username)

}
