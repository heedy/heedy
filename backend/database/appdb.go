package database

import (
	"errors"
	"fmt"
	"strings"
)

var ErrUnimplemented = errors.New("The given functionality is currently unimplemented")

type AppDB struct {
	adb *AdminDB
	c   *App
}

func NewAppDB(adb *AdminDB, c *App) *AppDB {
	return &AppDB{
		adb: adb,
		c:   c,
	}
}

// GetObjectAccess returns a ScopeArray that merges the current access
func (c *AppDB) GetObjectAccess(s *Object) (sa ScopeArray) {
	// It is assumed that s was retreived by calling ReadObject on a UserDB
	// and that s.Access holds the user's access permissions

	// First, we check if maybe we have full access to the object, which would make life so easy
	// If we own the object, we don't need to look at the user's access, since ours is *more*
	if s.App != nil && *s.App == c.c.ID {

		// If we have full access, we give full access
		if c.c.Scope.HasScope("self.objects") || c.c.Scope.HasScope("self.objects."+*s.Type) {
			sa.Scope = []string{"*"}
			sa.Update()
			return
		}

		// We do not have full access, so we list out the access we *do* have,
		hasAccess := []string{}
		for k := range c.c.Scope.scopeMap {
			if strings.HasPrefix(k, "self.objects:") || strings.HasPrefix(k, "self.objects."+*s.Type+":") {
				scopea := strings.SplitN(k, ":", 2)
				if len(scopea[1]) > 0 {
					hasAccess = append(hasAccess, scopea[1])
				}

			}
		}

		sa.Scope = hasAccess
		sa.Update()
		return
	}

	// OK, this means that the object is either belonging to our owner, or is shared with the owner.
	// Check which it is
	access := s.Access
	sprefix := "objects"
	if *s.Owner != *c.c.Owner {
		sprefix = "shared"
	} else if s.App == nil {
		// The app is nil, meaning that the object is totally owned by the user. The scope need to be replaced:
		access = *s.OwnerScope
	}

	// OK, we don't own it. Bummer. Maybe the object access list is *, in which case we can
	// check which scope we have
	if access.HasScope("*") {
		if c.c.Scope.HasScope(sprefix) || c.c.Scope.HasScope(sprefix+"."+*s.Type) {
			sa.Scope = []string{"*"}
			sa.Update()
			return
		}

		// Dammit, we don't have full access, so once again, list out the access we do have
		hasAccess := []string{}
		for k := range c.c.Scope.scopeMap {
			if strings.HasPrefix(k, sprefix+":") || strings.HasPrefix(k, sprefix+"."+*s.Type+":") {
				scopea := strings.SplitN(k, ":", 2)
				if len(scopea[1]) > 0 {
					hasAccess = append(hasAccess, scopea[1])
				}

			}
		}

		sa.Scope = hasAccess
		sa.Update()
		return
	}

	// Now we both don't own it, and we don't have full access. For the access listed, find out which ones we have
	hasAccess := []string{}
	for _, v := range access.Scope {
		if c.c.Scope.HasScope(sprefix+":"+v) || c.c.Scope.HasScope(sprefix+"."+*s.Type+":"+v) {
			hasAccess = append(hasAccess, v)
		}
	}
	sa.Scope = hasAccess
	sa.Update()
	return
}

func (db *AppDB) AdminDB() *AdminDB {
	return db.adb
}

func (db *AppDB) ID() string {
	return *db.c.Owner + "/" + db.c.ID
}

func (db *AppDB) Type() DBType {
	return AppType
}

func (db *AppDB) User() (*User, error) {
	// Read the owner
	return db.ReadUser(*db.c.Owner, nil)
}

func (db *AppDB) CreateUser(u *User) error {
	return ErrAccessDenied("A app cannot create a user")
}

func (db *AppDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	// A app can read a user:
	//	if the user is its owner, and owner:read scope
	//	if the user can be read by the owner, and has users:read scope
	if db.c.Scope.HasScope("users:read") || name == *db.c.Owner && db.c.Scope.HasScope("owner:read") {
		return NewUserDB(db.adb, *db.c.Owner).ReadUser(name, o)
	}

	return nil, ErrAccessDenied("Insufficient access to read the given user")
}

// UpdateUser updates the given portions of a user
func (db *AppDB) UpdateUser(u *User) error {
	// A app can read a user:
	//	if the user is its owner, and owner:read scope
	//	if the user can be read by the owner, and has users:read scope
	if db.c.Scope.HasScope("users:update") || u.ID == *db.c.Owner && db.c.Scope.HasScope("owner:update") {
		return NewUserDB(db.adb, *db.c.Owner).UpdateUser(u)
	}

	return ErrAccessDenied("Insufficient access to update the given user")
}
func (db *AppDB) DelUser(name string) error {
	if db.c.Scope.HasScope("users:delete") || name == *db.c.Owner && db.c.Scope.HasScope("owner:delete") {
		return NewUserDB(db.adb, *db.c.Owner).DelUser(name)
	}

	return ErrAccessDenied("Insufficient access to delete the given user")
}

func (db *AppDB) ListUsers(o *ListUsersOptions) ([]*User, error) {
	return nil, ErrUnimplemented
}

// CanCreateObject returns whether the given object can be
func (db *AppDB) CanCreateObject(s *Object) error {
	_, _, err := objectCreateQuery(db.adb.Assets().Config, s)
	if err != nil {
		return err
	}
	if s.App != nil && *s.App != db.c.ID {
		return ErrAccessDenied("Can't create a object for a different app")
	}
	if !db.c.Scope.HasScope("self.objects:create") && !db.c.Scope.HasScope("self.objects."+*s.Type+":create") {
		return ErrAccessDenied("Insufficient access to create a object of this type")
	}
	return nil
}

// CreateObject creates the object.
func (db *AppDB) CreateObject(s *Object) (string, error) {
	if s.App == nil || *s.App == "self" {
		s.App = &db.c.ID
	}
	if s.LastModified != nil {
		return "", ErrAccessDenied("Last Modified for object is readonly")
	}
	if *s.App != db.c.ID {
		return "", ErrAccessDenied("Can't create a object for a different app")
	}
	if s.Owner != nil && *s.Owner != *db.c.Owner {
		return "", ErrAccessDenied("Can't create a object for a different user")
	}
	// Must not explicily specify the owner for now
	s.Owner = nil
	if s.Type == nil || !db.c.Scope.HasScope("self.objects:create") && !db.c.Scope.HasScope("self.objects."+*s.Type+":create") {
		return "", ErrAccessDenied("Insufficient access to create a object of this type")
	}
	return db.adb.CreateObject(s)
}

// ReadObject reads the given object if the user has sufficient permissions
func (db *AppDB) ReadObject(id string, o *ReadObjectOptions) (*Object, error) {
	s, err := NewUserDB(db.adb, *db.c.Owner).ReadObject(id, o)
	if err != nil {
		return nil, err
	}
	s.Access = db.GetObjectAccess(s)

	return s, nil
}

// UpdateObject allows editing a object
func (db *AppDB) UpdateObject(s *Object) error {
	if s.LastModified != nil {
		return ErrAccessDenied("Modification date of object is readonly")
	}
	curs, err := db.ReadObject(s.ID, &ReadObjectOptions{
		Icon: false,
	})
	if err != nil {
		return err
	}

	// Since apps have their own special way of handling access, we check permissions here
	// and manually perform the update.

	if s.Name != nil || s.Owner != nil || s.App != nil || s.OwnerScope != nil {
		if !curs.Access.HasScope("update") {
			return ErrNotFound
		}
	} else {
		if !curs.Access.HasScope("update") && !curs.Access.HasScope("update:basic") {
			return ErrNotFound
		}
	}

	sColumns, sValues, err := objectUpdateQuery(db.adb.Assets().Config, s, *curs.Type)
	if err != nil {
		return err
	}

	sValues = append(sValues, s.ID)

	result, err := db.adb.Exec(fmt.Sprintf("UPDATE objects SET %s WHERE id=?;", sColumns), sValues...)
	return GetExecError(result, err)
}

// Can only delete objects that belong to *us*
func (db *AppDB) DelObject(id string) error {
	curs, err := db.ReadObject(id, &ReadObjectOptions{
		Icon: false,
	})
	if err != nil {
		return err
	}

	if !curs.Access.HasScope("delete") {
		return ErrAccessDenied("Insufficient permissions to delete the object")
	}
	result, err := db.adb.Exec("DELETE FROM objects WHERE id=?;", id)
	return GetExecError(result, err)
}

func (db *AppDB) ShareObject(objectid, userid string, sa *ScopeArray) error {
	return ErrUnimplemented
}

func (db *AppDB) UnshareObjectFromUser(objectid, userid string) error {
	return ErrUnimplemented
}

func (db *AppDB) UnshareObject(objectid string) error {
	return ErrUnimplemented
}

func (db *AppDB) GetObjectShares(objectid string) (m map[string]*ScopeArray, err error) {
	return nil, ErrUnimplemented
}

// ListObjects lists the given objects
func (db *AppDB) ListObjects(o *ListObjectsOptions) ([]*Object, error) {
	if o != nil && o.App != nil && *o.App == "self" {
		o.App = &db.c.ID
	}
	s, err := NewUserDB(db.adb, *db.c.Owner).ListObjects(o)
	if err != nil {
		return nil, err
	}
	ns := []*Object{}
	for _, v := range s {
		v.Access = db.GetObjectAccess(v)
		if v.Access.HasScope("read") {
			ns = append(ns, v)
		}
	}

	return ns, nil
}

func (db *AppDB) CreateApp(c *App) (string, string, error) {
	return "", "", ErrUnimplemented
}
func (db *AppDB) ReadApp(cid string, o *ReadAppOptions) (*App, error) {
	if cid == "self" {
		cid = db.c.ID
	}
	if cid != db.c.ID {
		return nil, ErrAccessDenied("Can't read other apps")
	}
	return readApp(db.adb, cid, o, `SELECT * FROM apps WHERE owner=? AND id=?;`, *db.c.Owner, cid)
}
func (db *AppDB) UpdateApp(c *App) error {
	if c.ID == "self" {
		c.ID = db.c.ID
	}
	if c.ID != db.c.ID {
		return ErrAccessDenied("Can't modify other apps")
	}
	if c.Scope != nil {
		return ErrAccessDenied("Can't change own scope")
	}
	if c.Plugin != nil {
		return ErrAccessDenied("Cannot modify app plugin value")
	}
	return updateApp(db.adb, c, "id=?", c.ID)
}
func (db *AppDB) DelApp(cid string) error {
	return ErrUnimplemented
}
func (db *AppDB) ListApps(o *ListAppOptions) ([]*App, error) {
	return nil, ErrUnimplemented
}

func (db *AppDB) ReadUserPreferences(username string) (map[string]map[string]interface{}, error) {
	return nil, ErrUnimplemented
}

func (db *AppDB) UpdatePluginPreferences(username string, plugin string, preferences map[string]interface{}) error {
	return ErrUnimplemented
}
func (db *AppDB) ReadPluginPreferences(username string, plugin string) (map[string]interface{}, error) {
	return nil, ErrUnimplemented
}
