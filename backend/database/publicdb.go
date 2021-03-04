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
func (db *PublicDB) Type() DBType {
	return PublicType
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

// CanCreateObject returns whether the given object can be
func (db *PublicDB) CanCreateObject(s *Object) error {
	if s.Type == nil {
		return ErrBadQuery("No object type given")
	}
	if s.Name == nil {
		return ErrBadQuery("The object needs a name")
	}
	return ErrAccessDenied("must be logged in to create the object")
}

func (db *PublicDB) CreateObject(s *Object) (string, error) {
	return "", ErrAccessDenied("You must be logged in to create objects")
}

// ReadObject reads the given object if it is shared
func (db *PublicDB) ReadObject(id string, o *ReadObjectOptions) (*Object, error) {
	return readObject(db.adb, id, o, `SELECT objects.*,json_group_array(ss.scope) AS access FROM objects, user_object_scope AS ss 
		WHERE objects.id=? AND ss.user='public' AND ss.object=objects.id;`, id)
}

// UpdateObject allows editing a object
func (db *PublicDB) UpdateObject(s *Object) error {
	if s.LastModified != nil {
		return ErrAccessDenied("Last Modified of object is readonly")
	}
	return updateObject(db.adb, s, `SELECT type,json_group_array(ss.scope) AS access FROM objects, user_object_scope AS ss
		WHERE objects.id=? AND ss.user='public' AND ss.object=objects.id;`, s.ID)
}

func (db *PublicDB) DelObject(id string) error {
	return ErrAccessDenied("You must be logged in to delete objects")
}

func (db *PublicDB) ShareObject(objectid, userid string, sa *ScopeArray) error {
	return ErrAccessDenied("You must be logged in to share objects")
}

func (db *PublicDB) UnshareObjectFromUser(objectid, userid string) error {
	return ErrAccessDenied("You must be logged in to delete object shares")
}

func (db *PublicDB) UnshareObject(objectid string) error {
	return ErrAccessDenied("You must be logged in to delete object shares")
}

func (db *PublicDB) GetObjectShares(objectid string) (m map[string]*ScopeArray, err error) {
	return nil, ErrAccessDenied("You must be logged in to get the object shares")
}

// ListObjects lists the given objects
func (db *PublicDB) ListObjects(o *ListObjectsOptions) ([]*Object, error) {
	return listObjects(db.adb, o, `SELECT objects.*,json_group_array(ss.scope) AS access FROM objects, user_object_scope AS ss
		WHERE %s AND ss.user='public' AND ss.object=objects.id GROUP BY objects.id %s;`)
}

func (db *PublicDB) CreateApp(c *App) (string, string, error) {
	return "", "", ErrAccessDenied("You must be logged in to create apps")
}
func (db *PublicDB) ReadApp(cid string, o *ReadAppOptions) (*App, error) {
	return nil, ErrAccessDenied("You must be logged in to read apps")
}
func (db *PublicDB) UpdateApp(c *App) error {
	return ErrAccessDenied("You must be logged in to update apps")
}
func (db *PublicDB) DelApp(cid string) error {
	return ErrAccessDenied("You must be logged in to delete apps")
}
func (db *PublicDB) ListApps(o *ListAppOptions) ([]*App, error) {
	return nil, ErrAccessDenied("You must be logged in to list apps")
}
func (db *PublicDB) ReadUserPreferences(username string) (map[string]map[string]interface{}, error) {
	return nil, ErrAccessDenied("You must be logged in to read preferences")
}
func (db *PublicDB) UpdatePluginPreferences(username string, plugin string, preferences map[string]interface{}) error {
	return ErrAccessDenied("You must be logged in to update preferences")
}
func (db *PublicDB) ReadPluginPreferences(username string, plugin string) (map[string]interface{}, error) {
	return nil, ErrAccessDenied("You must be logged in to read preferences")
}
