// Package users provides an API for managing user information.
package users

// BUG(joseph) This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"connectordb/streamdb/dbutil"
	"database/sql"
	"errors"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Standard Errors
	InvalidNameError    = errors.New("The provided name is not valid, it may not contain /, \\, space, ? or be blank")
	InvalidPointerError = errors.New("The provided pointer is nil")
	// statements

	READONLY_ERR = errors.New("Database is Read Only")

	ErrNothingToDelete = errors.New("The selected resource was not found, so it was not deleted.")
	ErrUserNotFound    = errors.New("The requested user was not found.")
	ErrDeviceNotFound  = errors.New("The requested device was not found.")
	ErrStreamNotFound  = errors.New("The requested stream was not found.")
)

type SqlUserDatabase struct {
	dbutil.SqlxMixin
	sqldb *sql.DB
}

func (db *SqlUserDatabase) InitSqlUserDatabase(sqldb *sql.DB, dbtype string) {
	db.InitSqlxMixin(sqldb, dbtype)
	db.sqldb = sqldb
}

// Checks to see if the name of a user/device/stream is legal.
func IsValidName(n string) bool {
	if strings.Contains(n, "/") ||
		strings.Contains(n, "\\") ||
		strings.Contains(n, " ") ||
		strings.Contains(n, "?") ||
		strings.Contains(n, "\t") ||
		strings.Contains(n, "\n") ||
		strings.Contains(n, "\r") ||
		strings.Contains(n, "#") ||
		len(n) == 0 {
		return false
	}

	return true
}

// Performs a set of tests on the result and error of a
// DELETE call to see what kind of error we should return.
func getDeleteError(result sql.Result, err error) error {
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNothingToDelete
	}

	return nil
}
