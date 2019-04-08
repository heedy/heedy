package database

import (
	"database/sql"
	"fmt"
)

func createUser(adb *AdminDB, u *User, sqlStatement string, args ...interface{}) error {
	// Only create the user if I have the users:create scope
	rows, err := adb.DB.Query(sqlStatement, args...)

	if err != nil {
		return err
	}
	canCreate := rows.Next()
	rows.Close()
	if !canCreate {
		return ErrAccessDenied("You do not have sufficient permissions to create users")
	}
	return adb.CreateUser(u)
}

func createStream(adb *AdminDB, s *Stream, sqlStatement string, args ...interface{}) (string, error) {
	// Only create the user if I have the users:create scope
	rows, err := adb.DB.Query(sqlStatement, args...)

	if err != nil {
		return "", err
	}
	canCreate := rows.Next()
	rows.Close()
	if !canCreate {
		return "", ErrAccessDenied("You do not have sufficient permissions to create a stream here")
	}
	return adb.CreateStream(s)
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

func readStream(adb *AdminDB, id string, o *ReadStreamOptions, selectStatement string, args ...interface{}) (*Stream, error) {
	s := &Stream{}
	err := adb.Get(s, selectStatement, args...)

	if err == sql.ErrNoRows {
		return nil, ErrAccessDenied("Either the stream does not exist, or you can't access it")
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

// updateStream updates the stream if the scopeSQL returns a result
func updateStream(adb *AdminDB, s *Stream, scopeSQL string, args ...interface{}) error {
	sColumns, sValues, err := streamUpdateQuery(s)
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
		return ErrAccessDenied("You do not have sufficient access to modify this stream")
	}

	sValues = append(sValues, s.ID)

	result, err := tx.Exec(fmt.Sprintf("UPDATE streams SET %s WHERE id=?;", sColumns), sValues...)
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

func delStream(adb *AdminDB, id string, sqlStatement string, args ...interface{}) error {
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
