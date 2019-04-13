package database

import (
	"database/sql"
	"fmt"
)

func createSource(adb *AdminDB, s *Source, sqlStatement string, args ...interface{}) (string, error) {
	// Only create the user if I have the users:create scope
	rows, err := adb.DB.Query(sqlStatement, args...)

	if err != nil {
		return "", err
	}
	canCreate := rows.Next()
	rows.Close()
	if !canCreate {
		return "", ErrAccessDenied("You do not have sufficient permissions to create a source here")
	}
	return adb.CreateSource(s)
}

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

func readSource(adb *AdminDB, id string, o *ReadSourceOptions, selectStatement string, args ...interface{}) (*Source, error) {
	s := &Source{}
	err := adb.Get(s, selectStatement, args...)

	if err == sql.ErrNoRows {
		return nil, ErrAccessDenied("Either the source does not exist, or you can't access it")
	}
	if o == nil || !o.Avatar {
		s.Avatar = nil
	}
	return s, err
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

// updateSource updates the source if the scopeSQL returns a result
func updateSource(adb *AdminDB, s *Source, scopeSQL string, args ...interface{}) error {
	sColumns, sValues, err := sourceUpdateQuery(s)
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
		return ErrAccessDenied("You do not have sufficient access to modify this source")
	}

	sValues = append(sValues, s.ID)

	result, err := tx.Exec(fmt.Sprintf("UPDATE sources SET %s WHERE id=?;", sColumns), sValues...)
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

func delSource(adb *AdminDB, id string, sqlStatement string, args ...interface{}) error {
	result, err := adb.DB.Exec(sqlStatement, args...)
	return getExecError(result, err)
}

// readUserScopes should give an error if the user doesn't exist
func readUserScopes(adb *AdminDB, username string, sqlStatement string, args ...interface{}) ([]string, error) {
	rows, err := adb.DB.Query(sqlStatement, args...)

	if err != nil {
		return nil, err
	}
	canRead := rows.Next()
	rows.Close()
	if !canRead {
		return nil, ErrAccessDenied("You do not have sufficient permissions to read this user's scopes")
	}
	var scopes []string
	err = adb.Select(&scopes, `SELECT DISTINCT(scope) FROM user_scopes WHERE user=?;`, username)

	return scopes, err
}
