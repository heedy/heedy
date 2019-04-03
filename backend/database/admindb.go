package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/heedy/heedy/backend/assets"

	"github.com/jmoiron/sqlx"
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
func (db *AdminDB) AuthUser(name string, password string) (string, string, error) {
	var selectResult struct {
		Name     string
		Password string
	}
	err := db.Get(&selectResult, "SELECT name,password FROM users WHERE name = ? LIMIT 1;", name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrUserNotFound
		}
		return "", "", err
	}
	if err = CheckPassword(password, selectResult.Password); err != nil {
		return "", "", ErrUserNotFound
	}
	return selectResult.Name, selectResult.Password, nil
}

// LoginToken gets an active login token's username
func (db *AdminDB) LoginToken(token string) (string, error) {
	var selectResult struct {
		User string
	}
	err := db.Get(&selectResult, "SELECT user FROM user_tokens WHERE token=?;", token)
	return selectResult.User, err
}

// AddLoginToken gets the token for a given user
func (db *AdminDB) AddLoginToken(user string) (token string, err error) {
	token, err = GenerateKey(15)
	if err != nil {
		return
	}
	result, err2 := db.Exec("INSERT INTO user_tokens (user,token) VALUES (?,?);", user, token)
	err = getExecError(result, err2)
	return
}

// RemoveLoginToken deletes the given token from the database
func (db *AdminDB) RemoveLoginToken(token string) error {
	result, err := db.Exec("DELETE FROM user_tokens WHERE token=?;", token)
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
	err := db.Get(u, "SELECT * FROM users WHERE name=?LIMIT 1;", name)

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
		result, err := db.DB.Exec(fmt.Sprintf("UPDATE users SET %s WHERE name=?;", userColumns), userValues...)
		return getExecError(result, err)

	}

	return ErrNoUpdate
}

// DelUser deletes the given user
func (db *AdminDB) DelUser(name string) error {
	// The user's group will be deleted by cascade on group owner
	result, err := db.Exec("DELETE FROM users WHERE name=?;", name)
	return getExecError(result, err)
}

// CreateGroup generates a group with the given owner groupID
func (db *AdminDB) CreateGroup(g *Group) (string, error) {
	groupColumns, groupValues, err := groupCreateQuery(g)
	if err != nil {
		return "", err
	}

	result, err := db.DB.Exec(fmt.Sprintf("INSERT INTO groups (%s) VALUES (%s);", groupColumns, qQ(len(groupValues))), groupValues...)
	err = getExecError(result, err)
	return g.ID, err
}

// ReadGroup reads a group by id
func (db *AdminDB) ReadGroup(id string, o *ReadGroupOptions) (*Group, error) {
	g := &Group{}
	err := db.Get(g, "SELECT * FROM groups WHERE (id=?) LIMIT 1;", id)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if o != nil && o.Avatar {
		g.Avatar = nil
	}

	return g, err
}

// UpdateGroup updates the given group (by ID)
func (db *AdminDB) UpdateGroup(g *Group) error {
	groupColumns, groupValues, err := groupUpdateQuery(g)
	if err != nil {
		return err
	}

	groupValues = append(groupValues, g.ID)

	// Allow updating groups that are not users
	result, err := db.Exec(fmt.Sprintf("UPDATE groups SET %s WHERE id=? AND id!=owner;", groupColumns), groupValues...)
	return getExecError(result, err)

}

// DelGroup deletes the given group. It does not permit deleting users.
func (db *AdminDB) DelGroup(id string) error {
	result, err := db.Exec("DELETE FROM groups WHERE id=? AND id!=owner;", id)
	return getExecError(result, err)
}

// CreateConnection creates a new connection. Nuff said.
func (db *AdminDB) CreateConnection(c *Connection) (string, string, error) {
	cColumns, cValues, err := connectionCreateQuery(c)
	if err != nil {
		return "", "", err
	}
	// id is last, apikey is second to last
	connectionid := c.ID
	apikey := *c.APIKey

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

	return connectionid, apikey, tx.Commit()

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

// GetConnectionByKey reads the connection corresponding to the given api key
func (db *AdminDB) GetConnectionByKey(apikey string) (*Connection, error) {
	if apikey == "" {
		return nil, ErrNotFound
	}
	c := &Connection{}
	err := db.Get(c, "SELECT * FROM connections WHERE (apikey=?) LIMIT 1;", apikey)
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

// CreateStream creates the stream
func (db *AdminDB) CreateStream(s *Stream) (string, error) {
	sColumns, sValues, err := streamCreateQuery(s)
	if err != nil {
		return "", err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO streams (%s) VALUES (%s);", sColumns, qQ(len(sValues))), sValues...)
	err = getExecError(result, err)

	return s.ID, err

}

// ReadStream gets the stream by ID
func (db *AdminDB) ReadStream(id string, o *ReadStreamOptions) (*Stream, error) {
	c := &Stream{}
	err := db.Get(c, "SELECT * FROM streams WHERE (id=?) LIMIT 1;", id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if o != nil && o.Avatar {
		c.Avatar = nil
	}
	return c, err
}

// UpdateStream updates the given stream by ID
func (db *AdminDB) UpdateStream(s *Stream) error {
	sColumns, sValues, err := streamUpdateQuery(s)
	if err != nil {
		return err
	}

	sValues = append(sValues, s.ID)

	// Allow updating groups that are not users
	result, err := db.Exec(fmt.Sprintf("UPDATE streams SET %s WHERE id=?;", sColumns), sValues...)
	return getExecError(result, err)

}

// DelStream deletes the given stream
func (db *AdminDB) DelStream(id string) error {
	result, err := db.Exec("DELETE FROM streams WHERE id=?;", id)
	return getExecError(result, err)
}

// AddUserScopes adds scopes to the user
func (db *AdminDB) AddUserScopeSets(username string, scopesets ...string) error {
	if username == "heedy" {
		return ErrAccessDenied
	}
	// Make sure users or public is not one of the scopesets
	for i := range scopesets {
		if scopesets[i] == "users" || scopesets[i] == "public" {
			scopesets[i] = scopesets[len(scopesets)-1]
			return db.AddUserScopeSets(username, scopesets[:len(scopesets)-1]...)
		}
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}

	for i := range scopesets {
		result, err := tx.Exec("INSERT OR IGNORE INTO user_scopesets(user,scopeset) VALUES (?,?);", username, scopesets[i])
		err = getExecError(result, err)
		if err != nil && err != ErrNotFound {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// RemUserScopeSets removes scope sets from a user, while ensuring that all the user's connections also lose the given scope sets
func (db *AdminDB) RemUserScopeSets(username string, scopesets ...string) error {
	if username == "heedy" {
		return ErrAccessDenied
	}
	for i := range scopesets {
		if scopesets[i] == "users" || scopesets[i] == "public" {
			return errors.New("bad_query: Cannot remove the 'users' or 'public' scopesets from a user")
		}
	}
	query, args, err := sqlx.In(`DELETE FROM user_scopesets WHERE user=? AND scopeset IN (?);`, username, scopesets)
	if err != nil {
		return err
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	result, err := tx.Exec(query, args...)
	err = getExecError(result, err)
	if err != nil && err != ErrNotFound {
		tx.Rollback()
		return nil
	}

	// Must also delete any scopes that the user no longer has from its connections
	result, err = tx.Exec(`DELETE FROM connection_scopes WHERE 
		connectionid IN (SELECT id FROM connections WHERE owner=?)
		AND scope NOT IN (SELECT scope FROM scopesets WHERE name IN (SELECT scopeset FROM user_scopesets WHERE user=?));
	`, username, username)
	err = getExecError(result, err)
	if err == ErrNotFound || err == nil {
		tx.Commit()
		return nil
	}
	tx.Rollback()
	return err
}

func (db *AdminDB) ReadUserScopeSets(username string) ([]string, error) {
	var scopesets []string
	err := db.Select(&scopesets, `SELECT scopeset FROM user_scopesets WHERE user=?;`, username)
	scopesets = append(scopesets, "users", "public")
	return scopesets, err
}

func (db *AdminDB) AddScopeSet(scopeset string, scopes ...string) error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}

	for i := range scopes {
		result, err := tx.Exec("INSERT OR IGNORE INTO scopesets(name,scope) VALUES (?,?);", scopeset, scopes[i])
		err = getExecError(result, err)
		if err != nil && err != ErrNotFound {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// RemScopeSet removes scopes from a set, while ensuring that all the user's connections also lose the scopes
func (db *AdminDB) RemScopeSet(scopeset string, scope ...string) error {
	query, args, err := sqlx.In(`DELETE FROM scopesets WHERE name=? AND scope IN (?);`, scopeset, scope)
	if err != nil {
		return err
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	result, err := tx.Exec(query, args...)
	err = getExecError(result, err)
	if err != nil && err != ErrNotFound {
		tx.Rollback()
		return nil
	}

	// Must also delete any scopes that were removed from all connections that were using it
	// Here we get once again thrown against the limitations of sqlite, and need to do some clever query writing instead of a simple join and IN over tuples.
	// We delete all values from connections scope where the corresponding connection's owner does not have a scope that is in connection_scopes.
	// WARNING: this probably does a full table scan.
	result, err = tx.Exec(`DELETE FROM connection_scopes WHERE 
		EXISTS (
			SELECT 1 FROM connections WHERE connection_scopes.connectionid=connections.id 
			AND connection_scopes.scope NOT IN (
				SELECT scope FROM scopesets WHERE name IN (SELECT scopeset FROM user_scopesets WHERE user=connections.owner
			)
		);
	`)
	err = getExecError(result, err)
	if err != ErrNotFound && err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (db *AdminDB) ReadScopeSet(scopeset string) ([]string, error) {
	var scopes []string
	err := db.Select(&scopes, `SELECT scope FROM scopesets WHERE name=?;`, scopeset)
	return scopes, err
}

func (db *AdminDB) GetAllScopeSets() (map[string][]string, error) {
	var s []struct {
		Scope string
		Name  string
	}
	var setnames []string
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	err = tx.Select(&s, `SELECT scope,name FROM scopesets;`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// This handles scopesets that have no actual scopes in them
	err = tx.Select(&setnames, `SELECT DISTINCT(scopeset) FROM user_scopesets;`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	var res = make(map[string][]string)
	for _, v := range s {
		if _, ok := res[v.Name]; !ok {
			res[v.Name] = []string{v.Scope}
		} else {
			res[v.Name] = append(res[v.Name], v.Scope)
		}
	}
	for _, v := range setnames {
		if _, ok := res[v]; !ok {
			res[v] = []string{}
		}
	}

	if _, ok := res["public"]; !ok {
		res["public"] = []string{}
	}
	if _, ok := res["users"]; !ok {
		res["users"] = []string{}
	}

	return res, tx.Commit()
}

// DeleteScopeSet deletes all references to a scopeset
func (db *AdminDB) DeleteScopeSet(scopeset string) error {

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	result, err := tx.Exec(`DELETE FROM scopesets WHERE name=?;`, scopeset)
	err = getExecError(result, err)
	if err != nil && err != ErrNotFound {
		tx.Rollback()
		return err
	}
	result, err = tx.Exec(`DELETE FROM user_scopesets WHERE scopeset=?;`, scopeset)
	err = getExecError(result, err)
	if err != nil && err != ErrNotFound {
		tx.Rollback()
		return err
	}

	// Must also delete any scopes that were removed from all connections that were using it
	// Here we get once again thrown against the limitations of sqlite, and need to do some clever query writing instead of a simple join and IN over tuples.
	// We delete all values from connections scope where the corresponding connection's owner does not have a scope that is in connection_scopes.
	// WARNING: this probably does a full table scan.
	result, err = tx.Exec(`DELETE FROM connection_scopes WHERE 
		EXISTS (
			SELECT 1 FROM connections WHERE connection_scopes.connectionid=connections.id 
			AND connection_scopes.scope NOT IN (
				SELECT scope FROM scopesets WHERE name IN (SELECT scopeset FROM user_scopesets WHERE user=connections.owner)
			)
		);
	`)
	err = getExecError(result, err)
	if err != ErrNotFound && err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (db *AdminDB) ReadUserScopes(username string) ([]string, error) {
	var scopes []string
	err := db.Select(&scopes, `SELECT DISTINCT(scope) FROM scopesets WHERE name IN (SELECT scopeset FROM user_scopesets WHERE user=?) OR name IN ('users', 'public');`, username)

	return scopes, err
}

/*

// AddGroupScopes adds the given scopes to the group. It only adds the scoped that the owner also has, and gives an error if the owner
// does not have the necessary permissions
func (db *AdminDB) AddGroupScopes(groupid string, scopes ...string) error {
	query, args, err := sqlx.In("SELECT COUNT(scope) FROM group_scopes INNER JOIN groups ON group_scopes.groupid=groups.owner WHERE groups.id=? AND group_scopes.scope IN (?)", groupid, scopes)
	if err != nil {
		return err
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	var scopeCount int
	err = tx.Get(&scopeCount, query, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	if scopeCount != len(scopes) {
		// Wrong scope count. However, maybe we want to add scopes to a group belonging to heedy user - this needs to always succeed, since heedy user is special
		var username string
		err = tx.Get(&username, "SELECT owner FROM groups WHERE id=?;", groupid)
		if err != nil || username != "heedy" {
			tx.Rollback()
			return errors.New("access_denied: you cannot add a scope that the group's owner does not have")
		}
		// Username is heedy, we're fine
	}

	for i := range scopes {
		result, err := tx.Exec("INSERT OR IGNORE INTO group_scopes(groupid,scope) VALUES (?,?);", groupid, scopes[i])
		err = getExecError(result, err)
		if err != nil && err != ErrNotFound {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// RemGroupScopes removes scopes from a group
func (db *AdminDB) RemGroupScopes(groupid string, scopes ...string) error {
	query, args, err := sqlx.In("DELETE FROM group_scopes WHERE groupid=? AND scope IN (?) ", groupid, scopes)
	if err != nil {
		return err
	}
	result, err := db.Exec(query, args...)
	err = getExecError(result, err)
	if err == ErrNotFound {
		return nil
	}
	return err
}

// GetGroupScopes gets the scopes in a group. This is also the method used to get a single user's
// scopes without the addition of group membership
func (db *AdminDB) GetGroupScopes(groupid string) ([]string, error) {
	var scopes []string
	err := db.Select(&scopes, "SELECT scope FROM group_scopes WHERE groupid=?", groupid)
	return scopes, err
}

// GetUserScopes returns all of the scopes that the user has. This also includes scopes that
// it has inherited through group membership. Use GetGroupScopes to get just the scopes
// of the specific user
func (db *AdminDB) GetUserScopes(username string) ([]string, error) {
	var scopes []string
	err := db.Select(&scopes, `SELECT DISTINCT(scope) FROM group_scopes WHERE groupid IN (?, 'public', 'users') OR groupid IN (
			SELECT groupid FROM group_members WHERE username=?
		);`, username, username)
	return scopes, err
}

/*

// SetGroupPermissions sets the given permissions
func (db *AdminDB) SetGroupPermissions(g *GroupPermissions) error {
	if g.Target == "" || g.Actor == "" || g.Target == g.Actor {
		return ErrInvalidQuery
	}
	if !g.GroupRead && !g.GroupWrite && !g.GroupDelete && !g.AddStream && !g.AddChild && !g.ListStreams && !g.ListChildren && !g.ListShared && !g.StreamRead && !g.StreamWrite && !g.StreamDelete && !g.DataRead && !g.DataWrite && !g.DataRemove && !g.ActionWrite {
		// Want to set action with NO permissions, so we just remove it from the group permissions if it exists
		_, err := db.Exec("DELETE FROM group_permissions WHERE target=? AND actor=?;", g.Target, g.Actor)
		return err
	}

	result, err := db.NamedExec(`INSERT OR REPLACE INTO group_permissions VALUES (
		:Target,:Actor,
		:GroupRead,:GroupWrite,:GroupDelete,
		:AddStream,:AddChild,
		:ListStreams,:ListChildren,:ListShared,
		:StreamRead,:StreamWrite,:StreamDelete,
		:DataRead,:DataWrite,:DataRemove,:ActionWrite
		)`, g)
	return getExecError(result, err)
}

// GetGroupPermissions returns the explicit permissions for the given group.
func (db *AdminDB) GetGroupPermissions(target string) (map[string]*GroupPermissions, error) {
	var gp []*GroupPermissions

	err := db.Select(&gp, "SELECT * FROM group_permissions WHERE target=?", target)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*GroupPermissions)
	for i := range gp {
		result[gp[i].Actor] = gp[i]
	}
	return result, nil
}
*/
