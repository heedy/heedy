/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbutil

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type SqlxMixin struct {
	DB                    *sqlx.DB
	sqlxPreparedStmtCache map[string]*sqlx.Stmt
}

// Initializes a sqlx mixin
func (db *SqlxMixin) InitSqlxMixin(sqldb *sqlx.DB) {
	db.sqlxPreparedStmtCache = make(map[string]*sqlx.Stmt)
	db.DB = sqldb
}

// This function returns a prepared statement, or prepares one for the given query
// stores it and returns it
func (db *SqlxMixin) GetOrPrepare(query string) (*sqlx.Stmt, error) {
	var err error

	prepared, ok := db.sqlxPreparedStmtCache[query]

	if ok {
		return prepared, nil
	}

	// Convert to the correct binding type
	query = db.DB.Rebind(query)

	prepared, err = db.DB.Preparex(query)

	if err != nil {
		return prepared, err
	}

	db.sqlxPreparedStmtCache[query] = prepared
	return prepared, nil
}

/**
This is a wrapper for the Get done in sqlx, it does auto conversion to stored
procedures executes them, and does conversion to the given query style for the
given database.

Gets a single item from the DB, remember to add LIMIT 1 if the DB doesn't know
about the query being for a unique item.
**/
func (db *SqlxMixin) Get(dest interface{}, query string, args ...interface{}) error {
	prep, err := db.GetOrPrepare(query)

	if err != nil {
		return err
	}

	return prep.Get(dest, args...)
}

/**
This is a wrapper for the Select done in sqlx, it does auto conversion to stored
procedures executes them, and does conversion to the given query style for the
given database.
**/
func (db *SqlxMixin) Select(dest interface{}, query string, args ...interface{}) error {
	prep, err := db.GetOrPrepare(query)

	if err != nil {
		return err
	}

	return prep.Select(dest, args...)
}

/**
This is a wrapper for the Exec done in sqlx, it does auto conversion to stored
procedures executes them, and does conversion to the given query style for the
given database.
**/
func (db *SqlxMixin) Exec(query string, args ...interface{}) (sql.Result, error) {
	prep, err := db.GetOrPrepare(query)

	if err != nil {
		return nil, err
	}

	return prep.Exec(args...)
}
