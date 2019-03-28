package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// AdminDB holds the main database, with admin access
type AdminDB struct {
	SqlxCache
}

// Close closes the backend database
func (db *AdminDB) Close() error {
	return db.DB.Close()
}

func (db *AdminDB) ID() string {
	return "heedy" // An administrative database acts as heedy
}

// AuthUser returns the groupid and password hash for the given user, or an authentication error
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
	groupColumns, groupValues, userColumns, userValues, err := userCreateQuery(u)
	if err != nil {
		return err
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	// Insert into user needs to be first, as group uses user as owner.
	result, err := tx.Exec(fmt.Sprintf("INSERT INTO users (%s) VALUES (%s);", userColumns, qQ(len(userValues))), userValues...)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}
	result, err = tx.Exec(fmt.Sprintf("INSERT INTO groups (%s) VALUES (%s);", groupColumns, qQ(len(groupValues))), groupValues...)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// ReadUser reads a user
func (db *AdminDB) ReadUser(name string) (*User, error) {
	u := &User{}
	err := db.Get(u, "SELECT * FROM groups WHERE id=? AND owner=id LIMIT 1;", name)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return u, err
}

// UpdateUser updates the given portions of a user
func (db *AdminDB) UpdateUser(u *User) error {
	if err := ValidName(u.ID); err != nil {
		return err
	}
	if u.Owner != nil {
		u.Owner = nil
	}
	groupColumns, groupValues, userColumns, userValues, err := userUpdateQuery(u)
	if err != nil {
		return err
	}

	if u.Name != nil {
		// A name change changes the group's ID also. We need to manually handle this.
		groupValues = append(groupValues, *u.Name)
		groupColumns = groupColumns + ",id=?"

		// Unfortunately, sqlite doesn't support alter table foreign keys, so we need to manually update the user name,
		// rather than cascading id change to user name
		userValues = append(userValues, *u.Name)
		if len(userValues) > 1 {
			userColumns = userColumns + ",name=?"
		} else {
			userColumns = "name=?"
		}
	}

	userValues = append(userValues, u.ID)
	groupValues = append(groupValues, u.ID)

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}

	// This needs to be first, in case user name is modified - the query will use old name here, and the ID will be cascaded to group owners
	if len(userValues) > 1 {
		// This uses a join to make sure that the group is in fact an existing user
		result, err := tx.Exec(fmt.Sprintf("UPDATE users SET %s WHERE name=?;", userColumns), userValues...)
		err = getExecError(result, err)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(groupValues) > 1 { // we added name, so check if >1
		// This uses a join to make sure that the group is in fact an existing user
		result, err := tx.Exec(fmt.Sprintf("UPDATE groups SET %s WHERE id=?;", groupColumns), groupValues...)
		err = getExecError(result, err)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// DelUser deletes the given user
func (db *AdminDB) DelUser(name string) error {
	// The user's group will be deleted by cascade on group owner
	result, err := db.Exec("DELETE FROM users WHERE name=?;", name)
	return getExecError(result, err)
}

// SearchUsers finds the users matching the given query string, up to the chosen limit.
func (db *AdminDB) SearchUsers(query string, limit int) ([]*User, error) {
	var searchResult []*User
	var err error

	if query == "" {
		if limit > 0 {
			err = db.Select(&searchResult, "SELECT * FROM groups WHERE id=? and id=owner LIMIT ?;", limit)
		} else {
			err = db.Select(&searchResult, "SELECT * FROM groups WHERE id=? and id=owner;")
		}
		return searchResult, err
	}

	//db.Select(&searchResult,"",query,limit)
	return nil, errors.New("Search unimplemented")
}

// CreateGroup generates a group with the given owner groupID
func (db *AdminDB) CreateGroup(g *Group) (string, error) {
	groupColumns, groupValues, err := groupCreateQuery(g)
	if err != nil {
		return "", err
	}

	result, err := db.Exec(fmt.Sprintf("INSERT INTO groups (%s) VALUES (%s);", groupColumns, qQ(len(groupValues))), groupValues...)
	err = getExecError(result, err)

	// The last element of groupValues is guaranteed to be the ID string
	return groupValues[len(groupValues)-1].(string), err
}

// ReadGroup reads a group by id
func (db *AdminDB) ReadGroup(id string) (*Group, error) {
	g := &Group{}
	err := db.Get(g, "SELECT * FROM groups WHERE (id=?) LIMIT 1;", id)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
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

	result, err := db.Exec(fmt.Sprintf("INSERT INTO connections (%s) VALUES (%s);", cColumns, qQ(len(cValues))), cValues...)
	err = getExecError(result, err)

	// id is last, apikey is second to last
	return cValues[len(cValues)-1].(string), cValues[len(cValues)-2].(string), err

}

// ReadConnection gets the connection associated with the given API key
func (db *AdminDB) ReadConnection(id string) (*Connection, error) {
	c := &Connection{}
	err := db.Get(c, "SELECT * FROM connections WHERE (id=?) LIMIT 1;", id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
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

	// id is last,
	return sValues[len(sValues)-1].(string), err

}

// ReadStream gets the stream by ID
func (db *AdminDB) ReadStream(id string) (*Stream, error) {
	c := &Stream{}
	err := db.Get(c, "SELECT * FROM streams WHERE (id=?) LIMIT 1;", id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
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

// AddGroupScopes adds the given scopes to the group, without checking their validity
func (db *AdminDB) AddGroupScopes(groupid string, scopes ...string) error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return err
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

// GetGroupScopes gets the scopes in a group
func (db *AdminDB) GetGroupScopes(groupid string) ([]string, error) {
	var scopes []string
	err := db.Select(&scopes, "SELECT scope FROM group_scopes WHERE groupid=?", groupid)
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
