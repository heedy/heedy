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

// ReadSource gets the source by ID
func (db *PublicDB) ReadSource(id string, o *ReadSourceOptions) (*Source, error) {
	return readSource(db.adb, id, o, `SELECT * FROM sources WHERE id=? AND EXISTS
		(SELECT 1 FROM public_can_read_source WHERE source=? LIMIT 1);`, id, id)
}

// UpdateSource updates the given source by ID
func (db *PublicDB) UpdateSource(s *Source) error {

	return updateSource(db.adb, s, `SELECT 1 FROM scopesets WHERE scope='sources:edit' AND name='public'
		AND EXISTS (SELECT 1 FROM public_can_read_source WHERE source=?);`, s.ID)
}

// DelSource deletes the given source, so long as it doesn't belong to a connection
func (db *PublicDB) DelSource(id string) error {
	return ErrAccessDenied("You must be logged in to delete a source")
}
