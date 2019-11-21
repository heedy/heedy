package database

import (
	"database/sql"
	"fmt"
)

func readUser(adb *AdminDB, name string, o *ReadUserOptions, selectStatement string, args ...interface{}) (*User, error) {
	u := &User{}
	err := adb.Get(u, selectStatement, args...)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if o == nil || !o.Icon {
		u.Icon = nil
	}
	return u, err
}

// updateUser updates the user if the given scopeSQL returns a result
func updateUser(adb *AdminDB, u *User, scopeSQL string, args ...interface{}) error {
	userColumns, userValues, err := userUpdateQuery(u)
	if err != nil {
		return err
	}

	tx, err := adb.Beginx()
	if err != nil {
		return err
	}

	rows, err := tx.Query(scopeSQL, args...)

	if err != nil {
		return err
	}
	canEdit := rows.Next()
	rows.Close()
	if !canEdit {
		tx.Rollback()
		return ErrAccessDenied("You do not have sufficient access to edit this user")
	}

	result, err := tx.Exec(fmt.Sprintf("UPDATE users SET %s WHERE username=?;", userColumns), userValues...)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
func delUser(adb *AdminDB, name string, sqlStatement string, args ...interface{}) error {
	result, err := adb.Exec(sqlStatement, args...)
	err = getExecError(result, err)
	if err == nil {
		// When deleting a user, we also remove the user from the list of admins
		err = adb.Assets().RemAdmin(name)
	}
	return err
}

func readObject(adb *AdminDB, objectid string, o *ReadObjectOptions, selectStatement string, args ...interface{}) (*Object, error) {
	s := &Object{}
	err := adb.Get(s, selectStatement, args...)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if !s.Access.HasScope("read") {
		return nil, ErrNotFound
	}

	if o == nil || !o.Icon {
		s.Icon = nil
	}

	return s, err
}

func readApp(adb *AdminDB, cid string, o *ReadAppOptions, selectStatement string, args ...interface{}) (*App, error) {
	c := &App{}
	err := adb.Get(c, selectStatement, args...)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
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

// updateObject uses a select statement that returns the object type if editing is permitted
func updateObject(adb *AdminDB, s *Object, selectStatement string, args ...interface{}) error {
	// Get the object type and scopes
	var sv struct {
		Stype  string     `db:"type"`
		Access ScopeArray `db:"access"`
	}
	err := adb.Get(&sv, selectStatement, args...)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	// Now check which scope we require for the update to succeed
	if s.Name != nil || s.Owner != nil || s.App != nil || s.Scopes != nil {
		if !sv.Access.HasScope("update") {
			return ErrNotFound
		}
	} else {
		if !sv.Access.HasScope("update") && !sv.Access.HasScope("update:basic") {
			return ErrNotFound
		}
	}

	sColumns, sValues, err := objectUpdateQuery(adb.Assets().Config, s, sv.Stype)
	if err != nil {
		return err
	}

	sValues = append(sValues, s.ID)

	// Allow updating groups that are not users
	result, err := adb.Exec(fmt.Sprintf("UPDATE objects SET %s WHERE id=?;", sColumns), sValues...)
	return getExecError(result, err)
}

func updateApp(adb *AdminDB, c *App, whereStatement string, args ...interface{}) error {

	// TODO: need to check if app belongs to plugin, and determine if any of the fields are readonly

	cColumns, cValues, err := appUpdateQuery(c)
	cValues = append(cValues, args...)
	result, err := adb.Exec(fmt.Sprintf("UPDATE apps SET %s WHERE %s", cColumns, whereStatement), cValues...)
	return getExecError(result, err)
}

// Here db is different, since it calls unshare
func shareObject(db DB, objectid, username string, sa *ScopeArray, scopeSQL string, args ...interface{}) error {
	adb := db.AdminDB()
	if len(sa.Scopes) == 0 {
		return db.UnshareObjectFromUser(objectid, username)
	}

	if !sa.HasScope("read") {
		return ErrBadQuery("To share a object, it needs to have the read scope active")
	}

	tx, err := adb.DB.Beginx()
	if err != nil {
		return err
	}

	rows, err := tx.Query(scopeSQL, args...)

	if err != nil {
		return err
	}
	canShare := rows.Next()
	rows.Close()
	if !canShare {
		tx.Rollback()
		return ErrAccessDenied("You do not have sufficient access to share this object")
	}

	result, err := adb.Exec("INSERT OR REPLACE INTO shared_objects(username,objectid,scopes) VALUES (?,?,?);", username, objectid, sa)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()

}

func unshareObjectFromUser(adb *AdminDB, objectid, userid string, selectStatement string, args ...interface{}) error {
	res, err := adb.Exec(selectStatement, args...)
	return getExecError(res, err)
}

func unshareObject(adb *AdminDB, objectid, selectStatement string, args ...interface{}) error {
	res, err := adb.Exec(selectStatement, args...)
	return getExecError(res, err)
}

func getObjectShares(adb *AdminDB, objectid, selectStatement string, args ...interface{}) (m map[string]*ScopeArray, err error) {
	var res []struct {
		Username string
		Scopes   *ScopeArray
	}

	err = adb.Select(&res, selectStatement, args...)
	if err != nil {
		return nil, err
	}

	m = make(map[string]*ScopeArray)
	for _, v := range res {
		m[v.Username] = v.Scopes
	}

	return m, err
}

func listObjects(adb *AdminDB, o *ListObjectsOptions, selectStatement string, args ...interface{}) ([]*Object, error) {
	var res []*Object
	q, v, err := listObjectsQuery(o)
	if err != nil {
		return nil, err
	}

	v = append(v, args...)
	limitString := ""
	if o != nil && o.Limit != nil {
		limitString = fmt.Sprintf("LIMIT %d", *o.Limit)
	} else {
		// If no limit is given, use limit of 1000
		limitString = fmt.Sprintf("LIMIT %d", 1000)
	}
	qstring := fmt.Sprintf(selectStatement, q, limitString)

	err = adb.Select(&res, qstring, v...)
	if err != nil {
		return nil, err
	}

	// Clear icons if not needed
	if o == nil || !o.Icon {
		for r := range res {
			res[r].Icon = nil
		}
	}
	return res, nil
}

// TODO: Needs to be redone for plugin apps
func listApps(adb *AdminDB, o *ListAppOptions, selectStatement string, args ...interface{}) ([]*App, error) {
	var res []*App
	err := adb.Select(&res, selectStatement, args...)
	if err != nil {
		return nil, err
	}
	if o == nil || !o.Icon {
		for r := range res {
			res[r].Icon = nil
		}
	}
	if o == nil || !o.AccessToken {
		for _, r := range res {
			r.AccessToken = nil
		}
	}

	return res, nil
}
