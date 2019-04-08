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

func (db *PublicDB) CreateStream(s *Stream) (string, error) {
	if s.Owner == nil {
		if s.Connection == nil {
			return "", ErrBadQuery("You must specify either an owner or a connection to which the stream should belong")
		}
		// Create stream for a connection...
		return "", ErrAccessDenied("Only a connection can add streams to itself")
	}

	// Check if we are allowed to create the stream
	return createStream(db.adb, s, `SELECT 1 FROM scopesets WHERE scope='streams:create' AND name='public' 
		AND EXISTS (SELECT 1 FROM public_can_read_user WHERE user=?);`, *s.Owner)
}

// ReadStream gets the stream by ID
func (db *PublicDB) ReadStream(id string, o *ReadStreamOptions) (*Stream, error) {
	return readStream(db.adb, id, o, `SELECT * FROM streams WHERE id=? AND EXISTS
		(SELECT 1 FROM public_can_read_stream WHERE stream=? LIMIT 1);`, id, id)
}

// UpdateStream updates the given stream by ID
func (db *PublicDB) UpdateStream(s *Stream) error {
	if s.Actor != nil || s.Access != nil || s.External != nil || s.Schema != nil {
		// The user is trying to edit core stream properties. Disallow this if the stream belongs to a connection
		return updateStream(db.adb, s, `SELECT 1 FROM scopesets WHERE scope='streams:edit' AND name='public'
			AND EXISTS (SELECT 1 FROM public_can_read_stream WHERE stream=?)
			AND EXISTS (SELECT 1 FROM streams WHERE id=? AND connection IS NULL);`, s.ID, s.ID)
	}

	return updateStream(db.adb, s, `SELECT 1 FROM scopesets WHERE scope='streams:edit' AND name='public'
		AND EXISTS (SELECT 1 FROM public_can_read_stream WHERE stream=?);`, s.ID)
}

// DelStream deletes the given stream, so long as it doesn't belong to a connection
func (db *PublicDB) DelStream(id string) error {
	return delStream(db.adb, id, `DELETE FROM streams WHERE id=?
			AND EXISTS (SELECT 1 FROM public_can_read_stream WHERE stream=?)
			AND EXISTS (SELECT 1 FROM streams WHERE id=? AND connection IS NULL)
			AND EXISTS (SELECT 1 FROM scopesets WHERE scope='streams:delete' AND name='public');`, id, id, id)
}
