// Package users provides an API for managing user information.
package users

// BUG(joseph) This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"database/sql"
	"errors"
	"streamdb/dbutil"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// A black and qhite question mark
	DEFAULT_ICON = `iVBORw0KGgoAAAANSUhEUgAAAEAAAABAAQMAAACQp+OdAAAABlBMVEUAA
	AAAAAClZ7nPAAAAAXRSTlMAQObYZgAAAAFiS0dEAIgFHUgAAAAJcEhZcwAACxMAAAsTAQCanBgAAACVS
	URBVCjPjdGxDcQgDAVQRxSUHoFRMtoxWkZhBEoKK/+IsaNc0ElQxE8K3xhBtLa4Gj4YNQBFEYHxjwFRJ
	OBU7AAsZOgVWSEJR68bajSUoOjfoK07NkP+h/jAiI8g2WgGdqRx+jVa/r0P2cx9EPE2zduUVxv2NHs6n
	Q6Z0BZQaX3F4/0od3xvE2TCtOeOs12UQl6c5Quj42jQ5zt8GQAAAABJRU5ErkJggg==`
	DEFAULT_PASSWORD_HASH = "SHA512"
)

var (
	// Standard Errors
	ERR_EMAIL_EXISTS    = errors.New("A user already exists with this email")
	ERR_USERNAME_EXISTS = errors.New("A user already exists with this username")
	ERR_INVALID_PTR     = errors.New("The provided pointer is nil")

	// statements

	READONLY_ERR = errors.New("Database is Read Only")
)

type UserDatabase struct {
	dbutil.SqlxMixin
}

func (db *UserDatabase) InitUserDatabase(sqldb *sql.DB, dbtype string) {
	db.InitSqlxMixin(sqldb, dbtype)
}
