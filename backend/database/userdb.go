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

func (db *UserDB) Type() DBType {
	return UserType
}

// User returns the user that is logged in
func (db *UserDB) User() (*User, error) {
	return db.ReadUser(db.user, &ReadUserOptions{
		Icon: true,
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
	if name == db.user || db.isAdmin() {
		return db.adb.ReadUser(name, o)
	}
	return readUser(db.adb, name, o, `SELECT * FROM users WHERE username=? AND (public_read OR users_read) LIMIT 1;`, name)
}

// UpdateUser updates the given portions of a user
func (db *UserDB) UpdateUser(u *User) error {
	if u.ID == db.user || db.isAdmin() {
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

func (db *UserDB) ListUsers(o *ListUsersOptions) ([]*User, error) {
	if db.isAdmin() {
		return db.adb.ListUsers(o)
	}
	return nil, ErrUnimplemented
}

// CanCreateObject returns whether the given object can be
func (db *UserDB) CanCreateObject(s *Object) error {
	_, _, err := objectCreateQuery(db.adb.Assets().Config, s)
	if err != nil {
		return err
	}
	if s.Owner != nil {
		if *s.Owner != db.user {
			return ErrAccessDenied("Cannot create a object for another user")
		}
	}
	if s.App != nil {
		return ErrAccessDenied("Can't create a object for a app")
	}
	return nil
}

// CreateObject creates the object.
func (db *UserDB) CreateObject(s *Object) (string, error) {
	if s.App != nil {
		return "", ErrAccessDenied("You cannot create objects belonging to a app")
	}
	if s.ModifiedDate != nil {
		return "", ErrAccessDenied("Last Modified status of object is readonly")
	}
	if s.Owner == nil {
		// If no owner is specified, assume the current user
		s.Owner = &db.user
	}
	if *s.Owner != db.user {
		return "", ErrAccessDenied("Cannot create a object belonging to someone else")
	}
	return db.adb.CreateObject(s)
}

// ReadObject reads the given object if the user has sufficient permissions
func (db *UserDB) ReadObject(id string, o *ReadObjectOptions) (*Object, error) {
	return readObject(db.adb, id, o, `SELECT objects.*,json_group_array(ss.scope) AS access FROM objects, user_object_scope AS ss 
		WHERE objects.id=? AND ss.user IN (?,'public','users') AND ss.object=objects.id;`, id, db.user)
}

// UpdateObject allows editing a object
func (db *UserDB) UpdateObject(s *Object) error {
	if s.ModifiedDate != nil {
		return ErrAccessDenied("Modification date of object is readonly")
	}
	return updateObject(db.adb, s, `SELECT type,json_group_array(ss.scope) AS access FROM objects, user_object_scope AS ss
		WHERE objects.id=? AND ss.user IN (?,'public','users') AND ss.object=objects.id;`, s.ID, db.user)
}

// Can only delete objects that belong to *us*
func (db *UserDB) DelObject(id string) error {
	result, err := db.adb.Exec("DELETE FROM objects WHERE id=? AND owner=? AND app IS NULL;", id, db.user)
	return GetExecError(result, err)
}

func (db *UserDB) ShareObject(objectid, userid string, sa *ScopeArray) error {
	return shareObject(db, objectid, userid, sa, `SELECT 1 FROM objects WHERE owner=? AND id=?`, db.user, objectid)
}

func (db *UserDB) UnshareObjectFromUser(objectid, userid string) error {
	return unshareObjectFromUser(db.adb, objectid, userid, `DELETE FROM shared_objects WHERE objectid=? AND username=? 
		AND EXISTS (SELECT 1 FROM objects WHERE owner=? AND id=objectid)`, objectid, userid, db.user)
}

func (db *UserDB) UnshareObject(objectid string) error {
	return unshareObject(db.adb, objectid, `DELETE FROM shared_objects WHERE objectid=?
		AND EXISTS (SELECT 1 FROM objects WHERE owner=? AND id=objectid)`, objectid, db.user)
}

func (db *UserDB) GetObjectShares(objectid string) (m map[string]*ScopeArray, err error) {
	return getObjectShares(db.adb, objectid, `SELECT username,scope FROM shared_objects WHERE objectid=?
		AND EXISTS (SELECT 1 FROM objects WHERE owner=? AND id=objectid)`, objectid, db.user)
}

// ListObjects lists the given objects
func (db *UserDB) ListObjects(o *ListObjectsOptions) ([]*Object, error) {
	if o != nil && o.Owner != nil && *o.Owner == "self" {
		o.Owner = &db.user
	}
	return listObjects(db.adb, o, `SELECT objects.*,json_group_array(ss.scope) AS access FROM objects, user_object_scope AS ss 
		WHERE %s AND ss.user IN (?,'public','users') AND ss.object=objects.id GROUP BY objects.id %s;`, db.user)
}

func (db *UserDB) CreateApp(c *App) (string, string, error) {

	if c.Owner == nil {
		// If no owner is specified, assume the current user
		c.Owner = &db.user
	}
	if *c.Owner != db.user {
		return "", "", ErrAccessDenied("Cannot create an app belonging to someone else")
	}
	if c.Plugin != nil {
		return "", "", ErrAccessDenied("Cannot create a plugin app")
	}
	return db.adb.CreateApp(c)
}
func (db *UserDB) ReadApp(cid string, o *ReadAppOptions) (*App, error) {
	// Can only read apps that belong to us
	return readApp(db.adb, cid, o, `SELECT * FROM apps WHERE owner=? AND id=?;`, db.user, cid)
}
func (db *UserDB) UpdateApp(c *App) error {
	if c.Plugin != nil {
		return ErrAccessDenied("Cannot modify app plugin value")
	}
	if c.SettingsSchema != nil {
		return ErrAccessDenied("Cannot modify app settings schema - only the app itself can do that.")
	}
	return updateApp(db.adb, c, `id=? AND owner=?`, c.ID, db.user)
}
func (db *UserDB) DelApp(cid string) error {
	// Can only delete apps that are not plugin-generated, unless the plugin is no longer active
	result, err := db.adb.Exec("DELETE FROM apps WHERE id=? AND owner=?;", cid, db.user)
	return GetExecError(result, err)
}
func (db *UserDB) ListApps(o *ListAppOptions) ([]*App, error) {
	if o != nil && o.Owner != nil && *o.Owner != db.user && *o.Owner != "self" {
		return nil, ErrAccessDenied("Can only list your own apps")
	}
	if o == nil {
		o = &ListAppOptions{}
	}
	o.Owner = &db.user // Add the owning user constraint
	return db.adb.ListApps(o)
}

func (db *UserDB) ReadUserSettings(username string) (map[string]map[string]interface{}, error) {
	if username != db.user {
		return nil, ErrAccessDenied("Cannot read other users' settings.")
	}
	return db.AdminDB().ReadUserSettings(username)
}

func (db *UserDB) UpdateUserPluginSettings(username string, plugin string, preferences map[string]interface{}) error {
	if username != db.user {
		return ErrAccessDenied("Cannot update other users' settings.")
	}
	return db.AdminDB().UpdateUserPluginSettings(username, plugin, preferences)
}
func (db *UserDB) ReadUserPluginSettings(username string, plugin string) (map[string]interface{}, error) {
	if username != db.user {
		return nil, ErrAccessDenied("Cannot read other users' settings.")
	}
	return db.AdminDB().ReadUserPluginSettings(username, plugin)
}
func (db *UserDB) ListUserSessions(username string) ([]UserSession, error) {
	if username != db.user {
		return nil, ErrAccessDenied("Cannot read other users' sessions.")
	}
	return db.adb.ListUserSessions(username)
}
func (db *UserDB) DelUserSession(username, id string) error {
	if username != db.user {
		return ErrAccessDenied("Cannot delete other users' sessions.")
	}
	return db.adb.DelUserSession(username, id)
}
