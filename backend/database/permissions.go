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

// given a user, performs a user update. It is given the sql that returns the distinct set of "scope LIKE 'user:edit%'" scopes
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

	// This uses a join to make sure that the group is in fact an existing user
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
