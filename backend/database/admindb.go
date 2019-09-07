package database

import (
	"database/sql"
	"fmt"

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

// User returns the user that is logged in
func (db *AdminDB) User() (*User, error) {
	return nil, nil
}

// AuthUser returns the user corresponding to the username and password, or an authentication error
func (db *AdminDB) AuthUser(username string, password string) (string, string, error) {
	var selectResult struct {
		UserName     string
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

// LoginToken gets an active login token's username
func (db *AdminDB) LoginToken(token string) (string, error) {
	var selectResult struct {
		UserName string
	}
	err := db.Get(&selectResult, "SELECT username FROM user_logintokens WHERE token=?;", token)
	return selectResult.UserName, err
}


// AddLoginToken gets the token for a given user
func (db *AdminDB) AddLoginToken(username string) (token string, err error) {
	token, err = GenerateKey(15)
	if err != nil {
		return
	}
	result, err2 := db.Exec("INSERT INTO user_logintokens (username,token) VALUES (?,?);", username, token)
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

	// Insert into user needs to be first, as group uses user as owner.
	result, err := db.DB.Exec(fmt.Sprintf("INSERT INTO users (%s) VALUES (%s);", userColumns, qQ(len(userValues))), userValues...)
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
	if o == nil || !o.Avatar {
		u.Avatar = nil
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
		result, err := db.DB.Exec(fmt.Sprintf("UPDATE users SET %s WHERE username=?;", userColumns), userValues...)
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

// CanCreateSource returns whether the given source can be
func (db *AdminDB) CanCreateSource(s *Source) error {
	_, _, err := sourceCreateQuery(db.Assets().Config, s)
	return err
}

// CreateSource creates the source
func (db *AdminDB) CreateSource(s *Source) (string, error) {
	sColumns, sValues, err := sourceCreateQuery(db.Assets().Config, s)
	if err != nil {
		return "", err
	}

	if s.Connection != nil {
		// We must insert while also setting the owner to the connection's owner
		sValues = append(sValues, *s.Connection)
		result, err := db.Exec(fmt.Sprintf("INSERT INTO sources (%s,owner) VALUES (%s,(SELECT owner FROM connections WHERE id=?));", sColumns, qQ(len(sValues)-1)), sValues...)
		err = getExecError(result, err)

		return s.ID, err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO sources (%s) VALUES (%s);", sColumns, qQ(len(sValues))), sValues...)
	err = getExecError(result, err)

	return s.ID, err

}

// ReadSource gets the source by ID
func (db *AdminDB) ReadSource(id string, o *ReadSourceOptions) (s *Source, err error) {
	s, err = readSource(db, id, o, `SELECT *,'["*"]' AS access FROM sources WHERE (id=?) LIMIT 1;`, id)
	return
}

// UpdateSource updates the given source by ID
func (db *AdminDB) UpdateSource(s *Source) error {
	return updateSource(db, s, `SELECT type,'["*"]' AS access FROM sources WHERE id=? LIMIT 1;`, s.ID)
}

// DelSource deletes the given source
func (db *AdminDB) DelSource(id string) error {
	result, err := db.Exec("DELETE FROM sources WHERE id=?;", id)
	return getExecError(result, err)
}

// ShareSource shares the given source with the given user, allowing the given set of scopes
func (db *AdminDB) ShareSource(sourceid, userid string, sa *ScopeArray) error {
	if len(sa.Scopes) == 0 {
		return db.UnshareSourceFromUser(sourceid, userid)
	}
	if !sa.HasScope("read") {
		return ErrBadQuery("To share a source, it needs to have the read scope active")
	}

	res, err := db.Exec("INSERT OR REPLACE INTO shared_sources(username,sourceid,scopes) VALUES (?,?,?);", userid, sourceid, sa)
	return getExecError(res, err)
}

// UnshareSourceFromUser Removes the given share from the source
func (db *AdminDB) UnshareSourceFromUser(sourceid, userid string) error {
	return unshareSourceFromUser(db, sourceid, userid, "DELETE FROM shared_sources WHERE sourceid=? AND username=?", sourceid, userid)
}

// UnshareSource deletes ALL the shares fro mthe source
func (db *AdminDB) UnshareSource(sourceid string) error {
	return unshareSource(db, sourceid, "DELETE FROM shared_sources WHERE sourceid=?", sourceid)
}

// GetSourceShares returns the shares of the source
func (db *AdminDB) GetSourceShares(sourceid string) (m map[string]*ScopeArray, err error) {
	return getSourceShares(db, sourceid, `SELECT username,scopes FROM shared_sources WHERE sourceid=?`, sourceid)
}

// ListSources lists the given sources
func (db *AdminDB) ListSources(o *ListSourcesOptions) ([]*Source,error) {
	return listSources(db,o,`SELECT *,'["*"]' AS access FROM sources WHERE %s %s;`)
}


// CreateConnection creates a new connection. Nuff said.
func (db *AdminDB) CreateConnection(c *Connection) (string, string, error) {
	cColumns, cValues, err := connectionCreateQuery(c)
	if err != nil {
		return "", "", err
	}
	// id is last, accessToken is second to last
	connectionid := c.ID
	accessToken := *c.AccessToken

	tx, err := db.DB.Beginx()
	if err != nil {
		return "", "", err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO connections (%s) VALUES (%s);", cColumns, qQ(len(cValues))), cValues...)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	scopes := db.Assets().Config.GetNewConnectionScopes()
	for i := range scopes {
		result, err := tx.Exec("INSERT INTO connection_scopes(connectionid,scope) VALUES (?,?);", connectionid, scopes[i])
		err = getExecError(result, err)
		if err != nil && err != ErrNotFound {
			tx.Rollback()
			return "", "", err
		}
	}

	return connectionid, accessToken, tx.Commit()

}

// ReadConnection gets the connection associated with the given API key
func (db *AdminDB) ReadConnection(id string, o *ReadConnectionOptions) (*Connection, error) {
	c := &Connection{}
	err := db.Get(c, "SELECT * FROM connections WHERE (id=?) LIMIT 1;", id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if o != nil && o.Avatar {
		c.Avatar = nil
	}
	return c, err
}

// GetConnectionByAccessToken reads the connection corresponding to the given access token
func (db *AdminDB) GetConnectionByAccessToken(accessToken string) (*Connection, error) {
	if accessToken == "" {
		return nil, ErrNotFound
	}
	c := &Connection{}
	err := db.Get(c, "SELECT * FROM connections WHERE (access_token=?) LIMIT 1;", accessToken)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return c, err
}

// UpdateConnection updates the given connection (by ID). Note that the inserted values will be written directly to
// the object.
func (db *AdminDB) UpdateConnection(c *Connection) error {
	cColumns, cValues, err := connectionUpdateQuery(c)
	if err != nil {
		return err
	}

	cValues = append(cValues, c.ID)

	// Allow updating groups that are not users
	result, err := db.Exec(fmt.Sprintf("UPDATE connections SET %s WHERE id=?;", cColumns), cValues...)
	return getExecError(result, err)

}

// DelConnection deletes the given connection.
func (db *AdminDB) DelConnection(id string) error {
	result, err := db.Exec("DELETE FROM connections WHERE id=?;", id)
	return getExecError(result, err)
}

// ListConnections lists connections
func (db *AdminDB) ListConnections(o *ListConnectionOptions) ([]*Connection,error) {
	return nil,ErrUnimplemented
}