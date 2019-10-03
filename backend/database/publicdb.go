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
	return "public" // The public db acts publically
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
	return readUser(db.adb, name, o, `SELECT * FROM users WHERE username=? AND public_read;`, name)

}

func (db *PublicDB) UpdateUser(u *User) error {
	return ErrAccessDenied("You must be logged in to update your user")
}

func (db *PublicDB) DelUser(name string) error {
	return ErrAccessDenied("You must be logged in to delete your user")
}

func (db *PublicDB) ListUsers(o *ListUsersOptions) ([]*User, error) {
	return nil, ErrUnimplemented
}

// CanCreateSource returns whether the given source can be
func (db *PublicDB) CanCreateSource(s *Source) error {
	if s.Type == nil {
		return ErrBadQuery("No source type given")
	}
	if s.Name == nil {
		return ErrBadQuery("The source needs a name")
	}
	return ErrAccessDenied("must be logged in to create the source")
}

func (db *PublicDB) CreateSource(s *Source) (string, error) {
	return "", ErrAccessDenied("You must be logged in to create sources")
}

// ReadSource reads the given source if it is shared
func (db *PublicDB) ReadSource(id string, o *ReadSourceOptions) (*Source, error) {
	return readSource(db.adb, id, o, `SELECT sources.*,json_group_array(ss.scope) AS access FROM sources, user_source_scopes AS ss 
		WHERE sources.id=? AND ss.user='public' AND ss.source=sources.id;`, id)
}

// UpdateSource allows editing a source
func (db *PublicDB) UpdateSource(s *Source) error {
	if s.NonEmpty != nil {
		return ErrAccessDenied("Empty status of source is readonly")
	}
	return updateSource(db.adb, s, `SELECT type,json_group_array(ss.scope) AS access FROM sources, user_source_scopes AS ss
		WHERE sources.id=? AND ss.user='public' AND ss.source=sources.id;`, s.ID)
}

func (db *PublicDB) DelSource(id string) error {
	return ErrAccessDenied("You must be logged in to delete sources")
}

func (db *PublicDB) ShareSource(sourceid, userid string, sa *ScopeArray) error {
	return ErrAccessDenied("You must be logged in to share sources")
}

func (db *PublicDB) UnshareSourceFromUser(sourceid, userid string) error {
	return ErrAccessDenied("You must be logged in to delete source shares")
}

func (db *PublicDB) UnshareSource(sourceid string) error {
	return ErrAccessDenied("You must be logged in to delete source shares")
}

func (db *PublicDB) GetSourceShares(sourceid string) (m map[string]*ScopeArray, err error) {
	return nil, ErrAccessDenied("You must be logged in to get the source shares")
}

// ListSources lists the given sources
func (db *PublicDB) ListSources(o *ListSourcesOptions) ([]*Source, error) {
	return listSources(db.adb, o, `SELECT sources.*,json_group_array(ss.scope) AS access FROM sources, user_source_scopes AS ss
		WHERE %s AND ss.user='public' AND ss.source=sources.id GROUP BY sources.id %s;`)
}

func (db *PublicDB) CreateConnection(c *Connection) (string, string, error) {
	return "", "", ErrAccessDenied("You must be logged in to create connections")
}
func (db *PublicDB) ReadConnection(cid string, o *ReadConnectionOptions) (*Connection, error) {
	return nil, ErrAccessDenied("You must be logged in to read connections")
}
func (db *PublicDB) UpdateConnection(c *Connection) error {
	return ErrAccessDenied("You must be logged in to update connections")
}
func (db *PublicDB) DelConnection(cid string) error {
	return ErrAccessDenied("You must be logged in to delete connections")
}
func (db *PublicDB) ListConnections(o *ListConnectionOptions) ([]*Connection, error) {
	return nil, ErrAccessDenied("You must be logged in to list connections")
}
