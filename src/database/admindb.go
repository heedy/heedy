package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AdminDB holds the main database, with admin access
type AdminDB struct {
	SqlxCache
}

// Close closes the backend database
func (db *AdminDB) Close() error {
	return db.DB.Close()
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
		return ErrInvalidQuery
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
		if len(userValues) > 0 {
			userColumns = userColumns + ","
		}
		userColumns = userColumns + "name=?"
	}

	groupValues = append(groupValues, u.ID)
	userValues = append(userValues, u.ID)

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
		result, err := tx.Exec(fmt.Sprintf("UPDATE groups SET %s WHERE id=? and id=owner;", groupColumns), groupValues...)
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
	if g.Owner == nil {
		// A group must have an owner
		return "", ErrInvalidQuery
	}
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
