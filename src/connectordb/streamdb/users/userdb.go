// Package users provides an API for managing user information.
package users

// BUG(joseph) This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"connectordb/streamdb/dbutil"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"strings"
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
	InvalidNameError    = errors.New("The provided name is not valid, it may not contain /, \\, space, ? or be blank")

	// statements

	READONLY_ERR = errors.New("Database is Read Only")
)

type UserDatabase struct {
	dbutil.SqlxMixin

	ukv = createClosure("a", "b")
}

func (db *UserDatabase) InitUserDatabase(sqldb *sql.DB, dbtype string) {
	db.InitSqlxMixin(sqldb, dbtype)
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
