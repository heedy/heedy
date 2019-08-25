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

// CanCreateSource returns whether the given source can be
func (db *UserDB) CanCreateSource(s *Source) error {
	_, _, err := sourceCreateQuery(db.adb.Assets().Config, s)
	if err != nil {
		return err
	}
	if s.Owner != nil {
		if *s.Owner != db.user {
			return ErrAccessDenied("Cannot create a source for another user")
		}
	}
	if s.Connection != nil {
		return ErrAccessDenied("Can't create a source for a connection")
	}
	return nil
}

// CreateSource creates the source.
func (db *UserDB) CreateSource(s *Source) (string, error) {
	if s.Connection != nil {
		return "", ErrAccessDenied("You cannot create sources belonging to a connection")
	}
	if s.Owner == nil {
		// If no owner is specified, assume the current user
		s.Owner = &db.user
	}
	if *s.Owner != db.user {
		return "", ErrAccessDenied("Cannot create a source belonging to someone else")
	}
	return db.adb.CreateSource(s)
}

// ReadSource reads the given source if the user has sufficient permissions
func (db *UserDB) ReadSource(id string, o *ReadSourceOptions) (*Source, error) {
	return readSource(db.adb, id, o, `SELECT sources.*,json_group_array(ss.scope) AS access FROM sources, user_source_scopes AS ss 
		WHERE sources.id=? AND ss.user IN (?,'public','users') AND ss.source=sources.id;`, id, db.user)
}

// UpdateSource allows editing a source
func (db *UserDB) UpdateSource(s *Source) error {
	return updateSource(db.adb, s, `SELECT type,json_group_array(ss.scope) AS access FROM sources, user_source_scopes AS ss
		WHERE sources.id=? AND ss.user IN (?,'public','users') AND ss.source=sources.id;`, s.ID, db.user)
}

// Can only delete sources that belong to *us*
func (db *UserDB) DelSource(id string) error {
	result, err := db.adb.Exec("DELETE FROM sources WHERE id=? AND owner=? AND connection IS NULL;", id, db.user)
	return getExecError(result, err)
}

func (db *UserDB) ShareSource(sourceid, userid string, sa *ScopeArray) error {
	return shareSource(db, sourceid, userid, sa, `SELECT 1 FROM sources WHERE owner=? AND id=?`, db.user, sourceid)
}

func (db *UserDB) UnshareSourceFromUser(sourceid, userid string) error {
	return unshareSourceFromUser(db.adb, sourceid, userid, `DELETE FROM shared_sources WHERE sourceid=? AND username=? 
		AND EXISTS (SELECT 1 FROM sources WHERE owner=? AND id=sourceid)`, sourceid, userid, db.user)
}

func (db *UserDB) UnshareSource(sourceid string) error {
	return unshareSource(db.adb, sourceid, `DELETE FROM shared_sources WHERE sourceid=?
		AND EXISTS (SELECT 1 FROM sources WHERE owner=? AND id=sourceid)`, sourceid, db.user)
}

func (db *UserDB) GetSourceShares(sourceid string) (m map[string]*ScopeArray, err error) {
	return getSourceShares(db.adb, sourceid, `SELECT username,scopes FROM shared_sources WHERE sourceid=?
		AND EXISTS (SELECT 1 FROM sources WHERE owner=? AND id=sourceid)`, sourceid, db.user)
}

// ListSources lists the given sources
func (db *UserDB) ListSources(o *ListSourcesOptions) ([]*Source, error) {
	return listSources(db.adb, o, `SELECT sources.*,json_group_array(ss.scope) AS access FROM sources, user_source_scopes AS ss 
		WHERE %s AND ss.user IN (?,'public','users') AND ss.source=sources.id GROUP BY sources.id %s;`, db.user)
}

func (db *UserDB) CreateConnection(c *Connection) (string, string, error) {

	if c.Owner == nil {
		// If no owner is specified, assume the current user
		c.Owner = &db.user
	}
	if *c.Owner != db.user {
		return "", "", ErrAccessDenied("Cannot create a connection belonging to someone else")
	}
	return db.adb.CreateConnection(c)
}
func (db *UserDB) ReadConnection(cid string, o *ReadConnectionOptions) (*Connection, error) {
	// Can only read connections that belong to us
	return readConnection(db.adb,cid,o,`SELECT * FROM connections WHERE owner=? AND id=?;`,db.user,cid)
}
func (db *UserDB) UpdateConnection(c *Connection) error {
	return updateConnection(db.adb,c,`id=? AND owner=?`,c.ID,db.user)
}
func (db *UserDB) DelConnection(cid string) error {
	result, err := db.adb.Exec("DELETE FROM connections WHERE id=? AND owner=?;", cid, db.user)
	return getExecError(result, err)
}
func (db *UserDB) ListConnections(o *ListConnectionOptions) ([]*Connection, error) {
	if o!=nil && o.User!=nil && *o.User!=db.user {
		return nil,ErrAccessDenied("Can only list your own connections")
	}
	return listConnections(db.adb,o,`SELECT * FROM connections WHERE owner=?`,db.user)
}
