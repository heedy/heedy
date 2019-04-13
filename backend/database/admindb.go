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
	err := db.Get(&selectResult, "SELECT user FROM user_logintokens WHERE token=?;", token)
	return selectResult.User, err
}

// AddLoginToken gets the token for a given user
func (db *AdminDB) AddLoginToken(user string) (token string, err error) {
	token, err = GenerateKey(15)
	if err != nil {
		return
	}
	result, err2 := db.Exec("INSERT INTO user_logintokens (user,token) VALUES (?,?);", user, token)
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

// CreateSource creates the source
func (db *AdminDB) CreateSource(s *Source) (string, error) {
	sColumns, sValues, err := sourceCreateQuery(s)
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
func (db *AdminDB) ReadSource(id string, o *ReadSourceOptions) (*Source, error) {
	c := &Source{}
	err := db.Get(c, "SELECT * FROM sources WHERE (id=?) LIMIT 1;", id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if o != nil && o.Avatar {
		c.Avatar = nil
	}
	return c, err
}

// UpdateSource updates the given source by ID
func (db *AdminDB) UpdateSource(s *Source) error {
	sColumns, sValues, err := sourceUpdateQuery(s)
	if err != nil {
		return err
	}

	sValues = append(sValues, s.ID)

	// Allow updating groups that are not users
	result, err := db.Exec(fmt.Sprintf("UPDATE sources SET %s WHERE id=?;", sColumns), sValues...)
	return getExecError(result, err)

}

// DelSource deletes the given source
func (db *AdminDB) DelSource(id string) error {
	result, err := db.Exec("DELETE FROM sources WHERE id=?;", id)
	return getExecError(result, err)
}
