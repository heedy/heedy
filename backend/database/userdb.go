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

// CreateStream creates the stream
func (db *UserDB) CreateStream(s *Stream) (string, error) {
	if s.Owner == nil {
		if s.Connection == nil {
			return "", ErrBadQuery("You must specify either an owner or a connection to which the stream should belong")
		}
		// Create stream for a connection...
		return "", ErrAccessDenied("Only a connection can add streams to itself")
	}
	if *s.Owner == db.user {
		// The owner is us, so we can directly create the stream
		return db.adb.CreateStream(s)
	}
	// The owner is someone else, so check if we are allowed to create a stream for them
	return createStream(db.adb, s, `SELECT 1 FROM user_scopes WHERE user=? AND scope='streams:create'
			AND EXISTS (SELECT 1 FROM user_can_read_user WHERE user=? AND target=?);`, db.user, db.user, *s.Owner)

}

// ReadStream gets the stream by ID
func (db *UserDB) ReadStream(id string, o *ReadStreamOptions) (*Stream, error) {
	return readStream(db.adb, id, o, `SELECT * FROM streams WHERE id=? AND EXISTS
		(SELECT 1 FROM user_can_read_stream WHERE stream=? AND user=? LIMIT 1);`, id, id, db.user)
}

// UpdateStream updates the given stream by ID
func (db *UserDB) UpdateStream(s *Stream) error {
	if s.Actor != nil || s.Access != nil || s.External != nil || s.Schema != nil {
		// The user is trying to edit core stream properties. Disallow this if the stream belongs to a connection
		return updateStream(db.adb, s, `SELECT 1 FROM user_can_read_stream WHERE stream=? AND user=? 
			AND EXISTS (
				SELECT 1 FROM streams WHERE id=? AND connection IS NULL AND (owner=? OR
						EXISTS (SELECT 1 FROM user_scopes WHERE user=? AND scope='streams:edit')
					)
				);`, s.ID, db.user, s.ID, db.user, db.user)
	}
	return updateStream(db.adb, s, `SELECT 1 FROM user_can_read_stream WHERE stream=? AND user=? 
		AND EXISTS (
			SELECT 1 FROM streams WHERE id=? AND (owner=? OR
					EXISTS (SELECT 1 FROM user_scopes WHERE user=? AND scope='streams:edit')
				)
			);`, s.ID, db.user, s.ID, db.user, db.user)
}

// DelStream deletes the given stream, so long as it doesn't belong to a connection
func (db *UserDB) DelStream(id string) error {
	return delStream(db.adb, id, `DELETE FROM streams WHERE id=?
			AND connection IS NULL
			AND (
				owner=? OR EXISTS (SELECT 1 FROM user_scopes WHERE user=? AND scope='streams:delete')
				AND EXISTS (SELECT 1 FROM user_can_read_stream WHERE stream=? AND user=?)
			);`, id, db.user, db.user, id, db.user)
}
