package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database/dbutil"
	"github.com/heedy/heedy/backend/events"
	"github.com/sirupsen/logrus"
)

// AdminDB holds the main database, with admin access
type AdminDB struct {
	SqlxCache

	a *assets.Assets
}

func (db *AdminDB) ReadPluginDatabaseVersion(plugin string) (int, error) {
	var curVersion int
	err := db.Get(&curVersion, `SELECT version FROM dbversion WHERE plugin=?`, plugin)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		curVersion = 0
	}
	return curVersion, nil
}

func (db *AdminDB) WritePluginDatabaseVersion(plugin string, version int) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO dbversion(plugin,version) VALUES (?,?)`, plugin, version)
	return err
}

// As allows performing a query with the given permissions level
func (db *AdminDB) As(identifier string) (DB, error) {
	if identifier == "heedy" {
		return db, nil
	}
	if identifier == "public" {
		return NewPublicDB(db), nil
	}
	// Now check if there is a slash in the identifier
	i := strings.Index(identifier, "/")

	username := identifier
	if i > -1 {
		username = identifier[:i]
		appid := identifier[i+1:]
		app, err := db.ReadApp(appid, nil)
		if err != nil {
			return nil, err
		}
		if *app.Owner != username {
			return nil, fmt.Errorf("User %s doesn't have app %s", username, appid)
		}
		return NewAppDB(db, app), nil
	}

	return NewUserDB(db, identifier), nil

}

// AdminDB returns the admin database
func (db *AdminDB) AdminDB() *AdminDB {
	return db
}

// Assets returns the assets being used for the database
func (db *AdminDB) Assets() *assets.Assets {
	return db.a
}

// Close closes the backend database
func (db *AdminDB) Close() error {
	return db.DB.Close()
}

func (db *AdminDB) ID() string {
	return "heedy" // An administrative database acts as heedy
}

func (db *AdminDB) Type() DBType {
	return AdminType
}

// User returns the user that is logged in
func (db *AdminDB) User() (*User, error) {
	return nil, nil
}

// AuthUser returns the user corresponding to the username and password, or an authentication error
func (db *AdminDB) AuthUser(username string, password string) (string, string, error) {
	var selectResult struct {
		UserName string
		Password string
	}
	err := db.Get(&selectResult, "SELECT username,password FROM users WHERE username = ? LIMIT 1;", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrUserNotFound
		}
		return "", "", err
	}
	if err = CheckPassword(password, selectResult.Password); err != nil {
		return "", "", ErrUserNotFound
	}
	return selectResult.UserName, selectResult.Password, nil
}

func shouldUpdateLastUsed(d dbutil.Date) bool {
	cy, cm, cd := time.Now().Date()
	dy, dm, dd := time.Time(d).Date()
	return cd > dd || cm > dm || cy > dy
}

// LoginToken gets an active login token's username, and sets the last acces date if not today
func (db *AdminDB) LoginToken(token string) (string, error) {
	var selectResult struct {
		UserName     string      `db:"username"`
		DateLastUsed dbutil.Date `db:"last_access_date"`
	}
	err := db.Get(&selectResult, "SELECT username,last_access_date FROM user_logintokens WHERE token=?;", token)
	if err == nil && shouldUpdateLastUsed(selectResult.DateLastUsed) {
		_, err = db.Exec("UPDATE user_logintokens SET last_access_date=DATE('now') WHERE token=?;", token)
	}
	return selectResult.UserName, err
}

// GetAppByAccessToken reads the app corresponding to the given access token,
// and sets the last access date if not today
func (db *AdminDB) GetAppByAccessToken(accessToken string) (*App, error) {
	if accessToken == "" {
		return nil, ErrNotFound
	}
	c := &App{}
	err := db.Get(c, "SELECT * FROM apps WHERE (access_token=?) LIMIT 1;", accessToken)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err == nil && (c.LastAccessDate == nil || shouldUpdateLastUsed(*c.LastAccessDate)) {
		_, err = db.Exec("UPDATE apps SET last_access_date=DATE('now') WHERE id=?;", c.ID)
	}
	return c, err
}

// AddLoginToken gets the token for a given user
func (db *AdminDB) AddLoginToken(username string, description string) (token string, err error) {
	token, err = GenerateKey(15)
	if err != nil {
		return
	}
	result, err2 := db.Exec("INSERT INTO user_logintokens (username,token,description) VALUES (?,?,?);", username, token, description)
	err = GetExecError(result, err2)
	return
}

// RemoveLoginToken deletes the given token from the database
func (db *AdminDB) RemoveLoginToken(token string) error {
	result, err := db.Exec("DELETE FROM user_logintokens WHERE token=?;", token)
	return GetExecError(result, err)
}

// CreateUser is the administrator version of create
func (db *AdminDB) CreateUser(u *User) error {
	userColumns, userValues, err := userCreateQuery(u)
	if err != nil {
		return err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO users (%s) VALUES (%s);", userColumns, QQ(len(userValues))), userValues...)
	return GetExecError(result, err)

}

// ReadUser reads a user
func (db *AdminDB) ReadUser(name string, o *ReadUserOptions) (*User, error) {
	u := &User{}
	err := db.Get(u, "SELECT * FROM users WHERE username=? LIMIT 1;", name)

	u.Password = nil

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if o == nil || !o.Icon {
		u.Icon = nil
	}

	return u, err
}

// UpdateUser updates the given portions of a user
func (db *AdminDB) UpdateUser(u *User) error {
	userColumns, userValues, err := userUpdateQuery(u)
	if err != nil {
		return err
	}

	// This needs to be first, in case user name is modified - the query will use old name here, and the ID will be cascaded
	if len(userValues) > 1 {
		// This uses a join to make sure that the group is in fact an existing user
		result, err := db.Exec(fmt.Sprintf("UPDATE users SET %s WHERE username=?;", userColumns), userValues...)
		err = GetExecError(result, err)
		if err == nil && u.UserName != nil {
			// The username was changed - make sure to update the configuration
			err = db.Assets().SwapAdmin(u.ID, *u.UserName)
		}
		return err

	}

	return ErrNoUpdate
}

// DelUser deletes the given user
func (db *AdminDB) DelUser(name string) error {
	// The user's group will be deleted by cascade on group owner
	result, err := db.Exec("DELETE FROM users WHERE username=?;", name)
	return GetExecError(result, err)
}

func (db *AdminDB) ListUsers(o *ListUsersOptions) (u []*User, err error) {
	err = db.Select(&u, "SELECT * FROM users WHERE username NOT IN ('heedy', 'users', 'public');")

	if o == nil || !o.Icon {
		for _, ui := range u {
			ui.Icon = nil
		}
	}
	return u, err
}

// CanCreateObject returns whether the given object can be
func (db *AdminDB) CanCreateObject(s *Object) error {
	_, _, err := objectCreateQuery(db.Assets().Config, s)
	return err
}

// CreateObject creates the object
func (db *AdminDB) CreateObject(s *Object) (string, error) {
	sColumns, sValues, err := objectCreateQuery(db.Assets().Config, s)
	if err != nil {
		return "", err
	}

	if s.App != nil {
		// We must insert while also setting the owner to the app's owner
		sValues = append(sValues, *s.App)
		result, err := db.Exec(fmt.Sprintf("INSERT INTO objects (%s,owner) VALUES (%s,(SELECT owner FROM apps WHERE id=?));", sColumns, QQ(len(sValues)-1)), sValues...)
		err = GetExecError(result, err)

		return s.ID, err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO objects (%s) VALUES (%s);", sColumns, QQ(len(sValues))), sValues...)
	err = GetExecError(result, err)

	return s.ID, err

}

// ReadObject gets the object by ID
func (db *AdminDB) ReadObject(id string, o *ReadObjectOptions) (s *Object, err error) {
	s, err = readObject(db, id, o, `SELECT *,'["*"]' AS access FROM objects WHERE (id=?) LIMIT 1;`, id)
	return
}

// UpdateObject updates the given object by ID
func (db *AdminDB) UpdateObject(s *Object) error {
	return updateObject(db, s, `SELECT type,'["*"]' AS access FROM objects WHERE id=? LIMIT 1;`, s.ID)
}

// DelObject deletes the given object
func (db *AdminDB) DelObject(id string) error {
	result, err := db.Exec("DELETE FROM objects WHERE id=?;", id)
	return GetExecError(result, err)
}

// ShareObject shares the given object with the given user, allowing the given set of scope
func (db *AdminDB) ShareObject(objectid, userid string, sa *ScopeArray) error {
	if len(sa.Scope) == 0 {
		return db.UnshareObjectFromUser(objectid, userid)
	}
	if !sa.HasScope("read") {
		return ErrBadQuery("To share a object, it needs to have the read scope active")
	}

	res, err := db.Exec("INSERT OR REPLACE INTO shared_objects(username,objectid,scope) VALUES (?,?,?);", userid, objectid, sa)
	return GetExecError(res, err)
}

// UnshareObjectFromUser Removes the given share from the object
func (db *AdminDB) UnshareObjectFromUser(objectid, userid string) error {
	return unshareObjectFromUser(db, objectid, userid, "DELETE FROM shared_objects WHERE objectid=? AND username=?", objectid, userid)
}

// UnshareObject deletes ALL the shares fro mthe object
func (db *AdminDB) UnshareObject(objectid string) error {
	return unshareObject(db, objectid, "DELETE FROM shared_objects WHERE objectid=?", objectid)
}

// GetObjectShares returns the shares of the object
func (db *AdminDB) GetObjectShares(objectid string) (m map[string]*ScopeArray, err error) {
	return getObjectShares(db, objectid, `SELECT username,scope FROM shared_objects WHERE objectid=?`, objectid)
}

// ListObjects lists the given objects
func (db *AdminDB) ListObjects(o *ListObjectsOptions) ([]*Object, error) {
	return listObjects(db, o, `SELECT *,'["*"]' AS access FROM objects WHERE %s %s;`)
}

// CreateApp creates a new app. Nuff said.
func (db *AdminDB) CreateApp(c *App) (string, string, error) {
	cColumns, cValues, err := appCreateQuery(c)
	if err != nil {
		return "", "", err
	}
	// id is last, accessToken is second to last
	appid := c.ID
	accessToken := *c.AccessToken

	tx, err := db.Beginx()
	if err != nil {
		return "", "", err
	}

	result, err := tx.Exec(fmt.Sprintf("INSERT INTO apps (%s) VALUES (%s);", cColumns, QQ(len(cValues))), cValues...)
	err = GetExecError(result, err)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	return appid, accessToken, tx.Commit()

}

// ReadApp gets the app associated with the given API key
func (db *AdminDB) ReadApp(aid string, o *ReadAppOptions) (*App, error) {
	return readApp(db, aid, o, "SELECT * FROM apps WHERE (id=?) LIMIT 1;", aid)
}

// UpdateApp updates the given app (by ID). Note that the inserted values will be written directly to
// the object.
func (db *AdminDB) UpdateApp(c *App) error {
	cColumns, cValues, err := appUpdateQuery(c)
	if err != nil {
		return err
	}

	cValues = append(cValues, c.ID)

	// Allow updating groups that are not users
	result, err := db.Exec(fmt.Sprintf("UPDATE apps SET %s WHERE id=?;", cColumns), cValues...)
	return GetExecError(result, err)

}

// DelApp deletes the given app.
func (db *AdminDB) DelApp(id string) error {
	result, err := db.Exec("DELETE FROM apps WHERE id=?;", id)
	return GetExecError(result, err)
}

// ListApps lists apps
func (db *AdminDB) ListApps(o *ListAppOptions) ([]*App, error) {
	a := []interface{}{}
	selectStmt := "SELECT * FROM apps"
	if o != nil && (o.Owner != nil || o.Plugin != nil) {
		selectStmt = selectStmt + " WHERE"
		if o.Owner != nil {
			selectStmt = selectStmt + " owner=?"
			a = append(a, *o.Owner)
		}
		if o.Plugin != nil {
			if o.Owner != nil {
				selectStmt = selectStmt + " AND"
			}
			if *o.Plugin == "" {
				selectStmt = selectStmt + " plugin IS NULL"
			} else {
				selectStmt = selectStmt + " plugin=?"
				a = append(a, *o.Plugin)
			}

		}
	}
	return listApps(db, o, selectStmt, a...)

}

// ReadUserSettings gets the given user's preferences. Returns default preferences if the user does not exist.
func (db *AdminDB) ReadUserSettings(username string) (map[string]map[string]interface{}, error) {
	var res []struct {
		Plugin string
		Key    string
		Value  []byte
	}

	// Start by constructing the result using default preference values from the configuration
	cfg := db.a.Config
	m := make(map[string]map[string]interface{})

	if len(cfg.UserSettingsSchema) > 0 {
		v := make(map[string]interface{})
		err := cfg.InsertUserSettingsDefaults(v)
		m["heedy"] = v
		if err != nil {
			return nil, err
		}
	}

	for _, p := range cfg.GetActivePlugins() {
		if len(cfg.Plugins[p].UserSettingsSchema) > 0 {
			v := make(map[string]interface{})
			err := cfg.Plugins[p].InsertUserSettingsDefaults(v)
			m[p] = v
			if err != nil {
				return nil, err
			}
		}
	}

	// Next, fill in settings that were updated by the user
	err := db.Select(&res, `SELECT plugin,key,value FROM user_settings WHERE user=?`, username)
	if err != nil {
		return nil, err
	}
	for _, resv := range res {
		var v interface{}
		err = json.Unmarshal(resv.Value, &v)
		if err != nil {
			return nil, err
		}
		m2, ok := m[resv.Plugin]
		if !ok {
			// There are settings for the plugin in the database despite there being no schema for them... This should be a warning
			logrus.Warnf("Existing settings found for plugin '%s', but no schema given.", resv.Plugin)
			m2 = make(map[string]interface{})
			m[resv.Plugin] = m2
		}
		m2[resv.Key] = v

	}

	return m, nil
}

func (db *AdminDB) ReadUserPluginSettings(username string, plugin string) (v map[string]interface{}, err error) {
	v = make(map[string]interface{})

	// First fill in the defaults
	cfg := db.a.Config
	if plugin == "heedy" {
		err = cfg.InsertUserSettingsDefaults(v)
	} else {
		pv, ok := cfg.Plugins[plugin]
		if !ok {
			return nil, errors.New("Unrecognized plugin")
		}
		err = pv.InsertUserSettingsDefaults(v)
	}
	if err != nil {
		return nil, err
	}

	var res []struct {
		Key   string
		Value []byte
	}

	// Next, fill in preferences that were updated by the user
	err = db.Select(&res, `SELECT key,value FROM user_settings WHERE user=? AND plugin=?`, username, plugin)
	if err != nil {
		return nil, err
	}
	for _, resv := range res {
		var v2 interface{}
		err = json.Unmarshal(resv.Value, &v2)
		if err != nil {
			return nil, err
		}
		v[resv.Key] = v2
	}

	return
}

func (db *AdminDB) UpdateUserPluginSettings(username string, plugin string, preferences map[string]interface{}) (err error) {
	if len(preferences) == 0 {
		return nil
	}

	cfg := db.a.Config
	if plugin == "heedy" {
		err = cfg.ValidateUserSettingsUpdate(preferences)
	} else {
		pv, ok := cfg.Plugins[plugin]
		if !ok {
			return errors.New("Unrecognized plugin")
		}
		err = pv.ValidateUserSettingsUpdate(preferences)
	}
	if err != nil {
		return err
	}

	// Now set the keys
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			// There was no error, so fire the event
			karray := make([]string, 0, len(preferences))
			for k := range preferences {
				karray = append(karray, k)
			}
			events.Fire(&events.Event{
				Event:  "user_settings_update",
				User:   username,
				Plugin: &plugin,
				Data: map[string]interface{}{
					"keys": karray,
				},
			})
		}
	}()

	for k, vi := range preferences {
		if vi != nil {
			var b []byte
			b, err = json.Marshal(vi)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`INSERT OR REPLACE INTO user_settings(user,plugin,key,value) VALUES (?,?,?,?);`, username, plugin, k, b)
		} else {
			// The value is nil, so we delete the element
			_, err = tx.Exec("DELETE FROM user_settings WHERE user=? AND plugin=? AND key=?", username, plugin, k)
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
