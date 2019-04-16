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
	return ErrAccessDenied("You must be logged in to create users")
}

func (db *PublicDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A user can be read if the user has public_read
	return readUser(db.adb, name, o, `SELECT * FROM users WHERE name=? AND public_read;`, name)

}

func (db *PublicDB) UpdateUser(u *User) error {
	return ErrAccessDenied("You must be logged in to update your user")
}

func (db *PublicDB) DelUser(name string) error {
	return ErrAccessDenied("You must be logged in to delete your user")
}

func (db *PublicDB) CreateSource(s *Source) (string, error) {
	return "", ErrAccessDenied("You must be logged in to create sources")
}
