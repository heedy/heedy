/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package users

// BUG(joseph) This should be moved to gorp once they support strong foreign key constraints
// right now we can't risk it without them

import (
	"database/sql"
	"dbsetup/dbutil"
	"errors"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/josephlewis42/multicache"
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

	nameValidator = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]*$")
)

type SqlUserDatabase struct {
	dbutil.SqlxMixin
	dbtype string
}

func (db *SqlUserDatabase) initSqlUserDatabase(sqldb *sqlx.DB) {
	db.InitSqlxMixin(sqldb)
	db.dbtype = sqldb.DriverName()
}

// Clear deletes all data stored in the userdb
func (db *SqlUserDatabase) Clear() {
	db.Exec("DELETE FROM Users;")
	db.Exec("DELETE FROM Devices;")
	db.Exec("DELETE FROM Streams;")
}

func NewUserDatabase(sqldb *sqlx.DB, cache bool, usersize int64, devsize int64, streamsize int64) UserDatabase {
	basedb := SqlUserDatabase{}
	basedb.initSqlUserDatabase(sqldb)

	if streamsize < 1 {
		streamsize = 1
	}

	streamCache, _ = multicache.NewDefaultMulticache(uint64(streamsize))

	if cache == false {
		return &basedb
	}

	// The cache sizes were already validated
	cached, _ := NewCacheMiddleware(&basedb, uint64(usersize), uint64(devsize), uint64(streamsize))

	return cached
}

// Checks to see if the name of a user/device/stream is legal.
func IsValidName(n string) bool {
	return nameValidator.MatchString(n) && len(n) > 0 && len(n) < 30
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
