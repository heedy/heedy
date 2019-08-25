package database

import "errors"

var ErrUnimplemented = errors.New("The given functionality is currently unimplemented")

type ConnectionDB struct {
	adb *AdminDB
	c   *Connection
}

func NewConnectionDB(adb *AdminDB, c *Connection) *ConnectionDB {
	return &ConnectionDB{
		adb: adb,
		c:   c,
	}
}

func (db *ConnectionDB) AdminDB() *AdminDB {
	return db.adb
}

func (db *ConnectionDB) ID() string {
	return db.c.ID
}

func (db *ConnectionDB) User() (*User, error) {
	return nil, nil
}

func (db *ConnectionDB) CreateUser(u *User) error {
	return ErrAccessDenied("Connections cannot currently create users")
}

func (db *ConnectionDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A connection can read a user:
	//	if the user is its owner, and owner:read scope
	//	if the user can be read by the owner, and has users:read scope
	return nil, ErrAccessDenied("Unimplemented")
}

// UpdateUser updates the given portions of a user
func (db *ConnectionDB) UpdateUser(u *User) error {
	// A connection can update a use:
	// 	if the user is its owner, and has owner:edit scope

	return ErrAccessDenied("You cannot modify other users")
}

func (db *ConnectionDB) DelUser(name string) error {
	return ErrAccessDenied("A connection cannot delete users")
}

// CanCreateSource returns whether the given source can be
func (db *ConnectionDB) CanCreateSource(s *Source) error {
	return ErrUnimplemented
}

// CreateSource creates the source.
func (db *ConnectionDB) CreateSource(s *Source) (string, error) {
	return "", ErrUnimplemented
}

// ReadSource reads the given source if the user has sufficient permissions
func (db *ConnectionDB) ReadSource(id string, o *ReadSourceOptions) (*Source, error) {
	return nil, ErrUnimplemented
}

// UpdateSource allows editing a source
func (db *ConnectionDB) UpdateSource(s *Source) error {
	return ErrUnimplemented
}

// Can only delete sources that belong to *us*
func (db *ConnectionDB) DelSource(id string) error {
	return ErrUnimplemented
}

func (db *ConnectionDB) ShareSource(sourceid, userid string, sa *ScopeArray) error {
	return ErrUnimplemented
}

func (db *ConnectionDB) UnshareSourceFromUser(sourceid, userid string) error {
	return ErrUnimplemented
}

func (db *ConnectionDB) UnshareSource(sourceid string) error {
	return ErrUnimplemented
}

func (db *ConnectionDB) GetSourceShares(sourceid string) (m map[string]*ScopeArray, err error) {
	return nil, ErrUnimplemented
}

// ListSources lists the given sources
func (db *ConnectionDB) ListSources(o *ListSourcesOptions) ([]*Source, error) {
	return nil, ErrUnimplemented
}

func (db *ConnectionDB) CreateConnection(c *Connection) (string, string, error) {
	return "", "", ErrAccessDenied("You must be logged in to create connections")
}
func (db *ConnectionDB) ReadConnection(cid string, o *ReadConnectionOptions) (*Connection, error) {
	return nil, ErrAccessDenied("You must be logged in to read connections")
}
func (db *ConnectionDB) UpdateConnection(c *Connection) error {
	return ErrAccessDenied("You must be logged in to update connections")
}
func (db *ConnectionDB) DelConnection(cid string) error {
	return ErrAccessDenied("You must be logged in to delete connections")
}
func (db *ConnectionDB) ListConnections(o *ListConnectionOptions) ([]*Connection, error) {
	return nil, ErrAccessDenied("You must be logged in to list connections")
}
