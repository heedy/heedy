package database

import (
	"strings"
	"errors"
)

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


// GetSourceAccess returns a ScopeArray that merges the current access
func (c *ConnectionDB) GetSourceAccess(s *Source) (sa ScopeArray) {
	// It is assumed that s was retreived by calling ReadSource on a UserDB
	// and that s.Access holds the user's access permissions

	// First, we check if maybe we have full access to the source, which would make life so easy
	// If we own the source, we don't need to look at the user's access, since ours is *more*
	if s.Connection!=nil && *s.Connection == c.c.ID {

		// If we have full access, we give full access
		if c.c.Scopes.HasScope("self.sources") || c.c.Scopes.HasScope("self.sources." + *s.Type) {
			sa.Scopes = []string{"*"}
			sa.Update()
			return
		}

		// We do not have full access, so we list out the access we *do* have, 
		hasAccess := []string{}
		for k := range c.c.Scopes.scopeMap {
			if strings.HasPrefix(k,"self.sources:") || strings.HasPrefix(k,"self.sources." + *s.Type + ":") {
				scopea := strings.SplitN(k,":",2)
				if len(scopea[1]) > 0 {
					hasAccess = append(hasAccess,scopea[1])
				}

			}
		}

		sa.Scopes = hasAccess
		sa.Update()
		return
	}

	// OK, this means that the source is either belonging to our owner, or is shared with the owner.
	// Check which it is 
	access := s.Access
	sprefix := "sources"
	if *s.Owner!=*c.c.Owner {
		sprefix = "shared"
	} else if s.Connection==nil {
		// The connection is nil, meaning that the source is totally owned by the user. The scopes need to be replaced:
		access = *s.Scopes
	}

	// OK, we don't own it. Bummer. Maybe the source access list is *, in which case we can
	// check which scopes we have
	if access.HasScope("*") {
		if c.c.Scopes.HasScope(sprefix) || c.c.Scopes.HasScope(sprefix+ "." + *s.Type) {
			sa.Scopes = []string{"*"}
			sa.Update()
			return
		}

		// Dammit, we don't have full access, so once again, list out the access we do have
		hasAccess := []string{}
		for k := range c.c.Scopes.scopeMap {
			if strings.HasPrefix(k,sprefix + ":") || strings.HasPrefix(k,sprefix + "." + *s.Type + ":") {
				scopea := strings.SplitN(k,":",2)
				if len(scopea[1]) > 0 {
					hasAccess = append(hasAccess,scopea[1])
				}

			}
		}

		sa.Scopes = hasAccess
		sa.Update()
		return
	}

	// Now we both don't own it, and we don't have full access. For the access listed, find out which ones we have
	hasAccess := []string{}
	for _,v := range access.Scopes {
		if c.c.Scopes.HasScope(sprefix + ":" + v) || c.c.Scopes.HasScope(sprefix +"."+ *s.Type + ":" + v) {
			hasAccess = append(hasAccess,v)
		}
	}
	sa.Scopes = hasAccess
	sa.Update()
	return
}

func (db *ConnectionDB) AdminDB() *AdminDB {
	return db.adb
}

func (db *ConnectionDB) ID() string {
	return *db.c.Owner + "/" + db.c.ID
}

func (db *ConnectionDB) User() (*User, error) {
	// Read the owner
	return db.ReadUser(*db.c.Owner,nil)
}

func (db *ConnectionDB) CreateUser(u *User) error {
	return ErrAccessDenied("A connection cannot create a user")
}

func (db *ConnectionDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A connection can read a user:
	//	if the user is its owner, and owner:read scope
	//	if the user can be read by the owner, and has users:read scope
	if db.c.Scopes.HasScope("users:read") || name == *db.c.Owner && db.c.Scopes.HasScope("owner:read") {
		return NewUserDB(db.adb,*db.c.Owner).ReadUser(name,o)
	}
	
	return nil, ErrAccessDenied("Insufficient access to read the given user")
}

// UpdateUser updates the given portions of a user
func (db *ConnectionDB) UpdateUser(u *User) error {
	// A connection can read a user:
	//	if the user is its owner, and owner:read scope
	//	if the user can be read by the owner, and has users:read scope
	if db.c.Scopes.HasScope("users:update") || u.ID == *db.c.Owner && db.c.Scopes.HasScope("owner:update") {
		return NewUserDB(db.adb,*db.c.Owner).UpdateUser(u)
	}
	
	return ErrAccessDenied("Insufficient access to update the given user")
}
func (db *ConnectionDB) DelUser(name string) error {
	if db.c.Scopes.HasScope("users:delete") || name == *db.c.Owner && db.c.Scopes.HasScope("owner:delete") {
		return NewUserDB(db.adb,*db.c.Owner).DelUser(name)
	}
	
	return ErrAccessDenied("Insufficient access to delete the given user")
}

// CanCreateSource returns whether the given source can be
func (db *ConnectionDB) CanCreateSource(s *Source) error {
	_, _, err := sourceCreateQuery(db.adb.Assets().Config, s)
	if err != nil {
		return err
	}
	if s.Connection != nil && *s.Connection!=db.c.ID {
		return ErrAccessDenied("Can't create a source for a different connection")
	}
	if !db.c.Scopes.HasScope("self.sources:create") && !db.c.Scopes.HasScope("self.sources." + *s.Type + ":create") {
		return ErrAccessDenied("Insufficient access to create a source of this type")
	}
	return nil
}

// CreateSource creates the source.
func (db *ConnectionDB) CreateSource(s *Source) (string, error) {
	if s.Connection==nil {
		s.Connection = &db.c.ID
	}
	if *s.Connection!=db.c.ID {
		return "",ErrAccessDenied("Can't create a source for a different connection")
	}
	if s.Owner!=nil && *s.Owner!=*db.c.Owner {
		return "",ErrAccessDenied("Can't create a source for a different user")
	}
	// Must not explicily specify the owner for now
	s.Owner = nil
	if s.Type==nil || !db.c.Scopes.HasScope("self.sources:create") && !db.c.Scopes.HasScope("self.sources." + *s.Type + ":create") {
		return "",ErrAccessDenied("Insufficient access to create a source of this type")
	}
	return db.adb.CreateSource(s)
}

// ReadSource reads the given source if the user has sufficient permissions
func (db *ConnectionDB) ReadSource(id string, o *ReadSourceOptions) (*Source, error) {
	s,err := NewUserDB(db.adb,*db.c.Owner).ReadSource(id,o)
	if err!=nil {
		return nil,err
	}
	s.Access = db.GetSourceAccess(s)
	
	return s,nil
}

// UpdateSource allows editing a source
func (db *ConnectionDB) UpdateSource(s *Source) error {
	curs,err := db.ReadSource(s.ID,&ReadSourceOptions{
		Avatar: false,
	})
	if err!=nil {
		return err
	}

	if s.Name != nil || s.Owner != nil || s.Connection != nil || s.Scopes != nil {
		if !curs.Access.HasScope("update") {
			return ErrNotFound
		}
	} else {
		if !curs.Access.HasScope("update") && !curs.Access.HasScope("update:basic") {
			return ErrNotFound
		}
	}

	return NewUserDB(db.adb,*db.c.Owner).UpdateSource(s)
}

// Can only delete sources that belong to *us*
func (db *ConnectionDB) DelSource(id string) error {
	curs,err := db.ReadSource(id,&ReadSourceOptions{
		Avatar: false,
	})
	if err!=nil {
		return err
	}

	if !curs.Access.HasScope("delete") {
		return ErrAccessDenied("Insufficient permissions to delete the source")
	}
	result, err := db.adb.Exec("DELETE FROM sources WHERE id=?;", id)
	return getExecError(result, err)
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
	if o!=nil && o.Connection!=nil && *o.Connection=="self" {
		o.Connection = &db.c.ID
	}
	s,err := NewUserDB(db.adb,*db.c.Owner).ListSources(o)
	if err!=nil {
		return nil,err
	}
	ns := []*Source{}
	for _,v := range s {
		v.Access = db.GetSourceAccess(v)
		if v.Access.HasScope("read") {
			ns = append(ns,v)
		}
	}

	return ns,nil
}

func (db *ConnectionDB) CreateConnection(c *Connection) (string, string, error) {
	return "", "", ErrUnimplemented
}
func (db *ConnectionDB) ReadConnection(cid string, o *ReadConnectionOptions) (*Connection, error) {
	return nil, ErrUnimplemented
}
func (db *ConnectionDB) UpdateConnection(c *Connection) error {
	return ErrUnimplemented
}
func (db *ConnectionDB) DelConnection(cid string) error {
	return ErrUnimplemented
}
func (db *ConnectionDB) ListConnections(o *ListConnectionOptions) ([]*Connection, error) {
	return nil, ErrUnimplemented
}
