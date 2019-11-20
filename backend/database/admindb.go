package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/heedy/heedy/backend/assets"
)

// AdminDB holds the main database, with admin access
type AdminDB struct {
	SqlxCache

	a *assets.Assets
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

func shouldUpdateLastUsed(d Date) bool {
	cy, cm, cd := time.Now().Date()
	dy, dm, dd := time.Time(d).Date()
	return cd > dd || cm > dm || cy > dy
}

// LoginToken gets an active login token's username, and sets the last acces date if not today
func (db *AdminDB) LoginToken(token string) (string, error) {
	var selectResult struct {
		UserName     string `db:"username"`
		DateLastUsed Date   `db:"last_access_date"`
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
	err = getExecError(result, err2)
	return
}

// RemoveLoginToken deletes the given token from the database
func (db *AdminDB) RemoveLoginToken(token string) error {
	result, err := db.Exec("DELETE FROM user_logintokens WHERE token=?;", token)
	return getExecError(result, err)
}

// CreateUser is the administrator version of create
func (db *AdminDB) CreateUser(u *User) error {
	userColumns, userValues, err := userCreateQuery(u)
	if err != nil {
		return err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO users (%s) VALUES (%s);", userColumns, QQ(len(userValues))), userValues...)
	return getExecError(result, err)

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

	// This needs to be first, in case user name is modified - the query will use old name here, and the ID will be cascaded to group owners
	if len(userValues) > 1 {
		// This uses a join to make sure that the group is in fact an existing user
		result, err := db.Exec(fmt.Sprintf("UPDATE users SET %s WHERE username=?;", userColumns), userValues...)
		return getExecError(result, err)

	}

	return ErrNoUpdate
}

// DelUser deletes the given user
func (db *AdminDB) DelUser(name string) error {
	// The user's group will be deleted by cascade on group owner
	result, err := db.Exec("DELETE FROM users WHERE username=?;", name)
	return getExecError(result, err)
}

func (db *AdminDB) ListUsers(o *ListUsersOptions) (u []*User, err error) {
	err = db.Select(&u, "SELECT * FROM users WHERE username NOT IN ('heedy', 'users', 'public');")

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
		err = getExecError(result, err)

		return s.ID, err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO objects (%s) VALUES (%s);", sColumns, QQ(len(sValues))), sValues...)
	err = getExecError(result, err)

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
	return getExecError(result, err)
}

// ShareObject shares the given object with the given user, allowing the given set of scopes
func (db *AdminDB) ShareObject(objectid, userid string, sa *ScopeArray) error {
	if len(sa.Scopes) == 0 {
		return db.UnshareObjectFromUser(objectid, userid)
	}
	if !sa.HasScope("read") {
		return ErrBadQuery("To share a object, it needs to have the read scope active")
	}

	res, err := db.Exec("INSERT OR REPLACE INTO shared_objects(username,objectid,scopes) VALUES (?,?,?);", userid, objectid, sa)
	return getExecError(res, err)
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
	return getObjectShares(db, objectid, `SELECT username,scopes FROM shared_objects WHERE objectid=?`, objectid)
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
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	scopes := db.Assets().Config.GetNewAppScopes()
	for i := range scopes {
		result, err := tx.Exec("INSERT INTO app_scopes(appid,scope) VALUES (?,?);", appid, scopes[i])
		err = getExecError(result, err)
		if err != nil && err != ErrNotFound {
			tx.Rollback()
			return "", "", err
		}
	}

	return appid, accessToken, tx.Commit()

}

// ReadApp gets the app associated with the given API key
func (db *AdminDB) ReadApp(id string, o *ReadAppOptions) (*App, error) {
	c := &App{}
	err := db.Get(c, "SELECT * FROM apps WHERE (id=?) LIMIT 1;", id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if o != nil && o.Icon {
		c.Icon = nil
	}
	if o == nil || !o.Icon {
		c.Icon = nil
	}
	if o == nil || !o.AccessToken {
		if c.AccessToken != nil {
			c.AccessToken = nil
		} else {
			// Make empty access token show up as empty, so services can know
			// that no access token is available
			if c.AccessToken == nil {
				emptyString := ""
				c.AccessToken = &emptyString
			}
		}

	} else {
		// Make empty access token show up as empty
		if c.AccessToken == nil {
			emptyString := ""
			c.AccessToken = &emptyString
		}
	}
	return c, err
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
	return getExecError(result, err)

}

// DelApp deletes the given app.
func (db *AdminDB) DelApp(id string) error {
	result, err := db.Exec("DELETE FROM apps WHERE id=?;", id)
	return getExecError(result, err)
}

// ListApps lists apps
func (db *AdminDB) ListApps(o *ListAppOptions) ([]*App, error) {
	var c []*App
	a := []interface{}{}
	selectStmt := "SELECT * FROM apps"
	if o != nil && (o.User != nil || o.Plugin != nil) {
		selectStmt = selectStmt + " WHERE"
		if o.User != nil {
			selectStmt = selectStmt + " owner=?"
			a = append(a, *o.User)
		}
		if o.Plugin != nil {
			if o.User != nil {
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
	err := db.Select(&c, selectStmt, a...)
	if err == nil && o != nil {
		if o.Icon != nil && *o.Icon == false {
			for _, cc := range c {
				cc.Icon = nil
			}
		}
	}
	return c, err
}
