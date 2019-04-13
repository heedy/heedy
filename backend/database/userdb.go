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

func (db *UserDB) isAdmin() bool {
	return db.adb.Assets().Config.UserIsAdmin(db.user)
}

func (db *UserDB) CreateUser(u *User) error {
	// Only an admin is allowed to create users
	if db.isAdmin() {
		return db.adb.CreateUser(u)
	}
	return ErrAccessDenied("You do not have sufficient permissions to create users")
}

func (db *UserDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A user can be read if it is the current user, OR if the user gave read access to itself
	if name == db.user {
		return db.adb.ReadUser(name, o)
	}
	return readUser(db.adb, name, o, `SELECT * FROM users WHERE name=? AND (public_read OR users_read) LIMIT 1;`, name)
}

// UpdateUser updates the given portions of a user
func (db *UserDB) UpdateUser(u *User) error {
	if u.ID == db.user {
		return db.adb.UpdateUser(u)
	}

	return ErrAccessDenied("You cannot modify other users")
}

func (db *UserDB) DelUser(name string) error {
	// A user can only delete themselves. If they are admins, they can delete any user
	if name == db.user || db.isAdmin() {
		return db.adb.DelUser(name)
	}

	return ErrAccessDenied("You cannot delete other users")
}

/*
// CreateSource creates the source
func (db *UserDB) CreateSource(s *Source) (string, error) {
	if s.Owner == nil {
		if s.Connection == nil {
			return "", ErrBadQuery("You must specify either an owner or a connection to which the source should belong")
		}
		// Create source for a connection...
		return "", ErrAccessDenied("Only a connection can add sources to itself")
	}
	if *s.Owner == db.user {
		// The owner is us, so we can directly create the source
		return db.adb.CreateSource(s)
	}
	// The owner is someone else, so check if we are allowed to create a source for them
	return createSource(db.adb, s, `SELECT 1 FROM user_scopes WHERE user=? AND scope='sources:create'
			AND EXISTS (SELECT 1 FROM user_can_read_user WHERE user=? AND target=?);`, db.user, db.user, *s.Owner)

}

// ReadSource gets the source by ID
func (db *UserDB) ReadSource(id string, o *ReadSourceOptions) (*Source, error) {
	return readSource(db.adb, id, o, `SELECT * FROM sources WHERE id=? AND EXISTS
		(SELECT 1 FROM user_can_read_source WHERE source=? AND user=? LIMIT 1);`, id, id, db.user)
}

// UpdateSource updates the given source by ID
func (db *UserDB) UpdateSource(s *Source) error {
	if s.Actor != nil || s.Access != nil || s.External != nil || s.Schema != nil {
		// The user is trying to edit core source properties. Disallow this if the source belongs to a connection
		return updateSource(db.adb, s, `SELECT 1 FROM user_can_read_source WHERE source=? AND user=?
			AND EXISTS (
				SELECT 1 FROM sources WHERE id=? AND connection IS NULL AND (owner=? OR
						EXISTS (SELECT 1 FROM user_scopes WHERE user=? AND scope='sources:edit')
					)
				);`, s.ID, db.user, s.ID, db.user, db.user)
	}
	return updateSource(db.adb, s, `SELECT 1 FROM user_can_read_source WHERE source=? AND user=?
		AND EXISTS (
			SELECT 1 FROM sources WHERE id=? AND (owner=? OR
					EXISTS (SELECT 1 FROM user_scopes WHERE user=? AND scope='sources:edit')
				)
			);`, s.ID, db.user, s.ID, db.user, db.user)
}

// DelSource deletes the given source, so long as it doesn't belong to a connection
func (db *UserDB) DelSource(id string) error {
	return delSource(db.adb, id, `DELETE FROM sources WHERE id=?
			AND connection IS NULL
			AND (
				owner=? OR EXISTS (SELECT 1 FROM user_scopes WHERE user=? AND scope='sources:delete')
				AND EXISTS (SELECT 1 FROM user_can_read_source WHERE source=? AND user=?)
			);`, id, db.user, db.user, id, db.user)
}

func (db *UserDB) ReadSourceData(id string, q *sources.Query) (sources.DatapointIterator, error) {
	return readSourceData(db.adb, id, q, `SELECT 1 FROM user_can_read_source WHERE source=? AND user=? LIMIT 1`, id, db.user)
}
*/
