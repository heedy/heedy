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
	if o == nil || !o.Avatar {
		u.Avatar = nil
	}
	return u, err
}

// updateUser updates the user if the given scopeSQL returns a result
func updateUser(adb *AdminDB, u *User, scopeSQL string, args ...interface{}) error {
	userColumns, userValues, err := userUpdateQuery(u)
	if err != nil {
		return err
	}

	tx, err := adb.DB.Beginx()
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

	result, err := tx.Exec(fmt.Sprintf("UPDATE users SET %s WHERE name=?;", userColumns), userValues...)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
func delUser(adb *AdminDB, name string, sqlStatement string, args ...interface{}) error {
	result, err := adb.DB.Exec(sqlStatement, args...)
	return getExecError(result, err)
}

func readSource(adb *AdminDB, sourceid string, o *ReadSourceOptions, selectStatement string, args ...interface{}) (*Source, error) {
	s := &Source{}
	err := adb.Get(s, selectStatement, args...)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if !s.Access.HasScope("read") {
		return nil, ErrNotFound
	}

	if o == nil || !o.Avatar {
		s.Avatar = nil
	}

	return s, err
}

// updateSource uses a select statement that returns the source type if editing is permitted
func updateSource(adb *AdminDB, s *Source, selectStatement string, args ...interface{}) error {
	// Get the source type and scopes
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
	if s.Name != nil || s.Owner != nil || s.Connection != nil || s.Scopes != nil {
		if !sv.Access.HasScope("update") {
			return ErrNotFound
		}
	} else {
		if !sv.Access.HasScope("update") && !sv.Access.HasScope("update:basic") {
			return ErrNotFound
		}
	}

	sColumns, sValues, err := sourceUpdateQuery(adb.Assets().Config, s, sv.Stype)
	if err != nil {
		return err
	}

	sValues = append(sValues, s.ID)

	// Allow updating groups that are not users
	result, err := adb.Exec(fmt.Sprintf("UPDATE sources SET %s WHERE id=?;", sColumns), sValues...)
	return getExecError(result, err)
}

// Here db is different, since it calls unshare
func shareSource(db DB, sourceid, username string, sa *ScopeArray, scopeSQL string, args ...interface{}) error {
	adb := db.AdminDB()
	if len(sa.Scopes) == 0 {
		return db.UnshareSourceFromUser(sourceid, username)
	}

	if !sa.HasScope("read") {
		return ErrBadQuery("To share a source, it needs to have the read scope active")
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
		return ErrAccessDenied("You do not have sufficient access to share this source")
	}

	result, err := adb.Exec("INSERT OR REPLACE INTO shared_sources(username,sourceid,scopes) VALUES (?,?,?);", username, sourceid, sa)
	err = getExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()

}

func unshareSourceFromUser(adb *AdminDB, sourceid, userid string, selectStatement string, args ...interface{}) error {
	res, err := adb.Exec(selectStatement, args...)
	return getExecError(res, err)
}

func unshareSource(adb *AdminDB, sourceid, selectStatement string, args ...interface{}) error {
	res, err := adb.Exec(selectStatement, args...)
	return getExecError(res, err)
}

func getSourceShares(adb *AdminDB, sourceid, selectStatement string, args ...interface{}) (m map[string]*ScopeArray, err error) {
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
