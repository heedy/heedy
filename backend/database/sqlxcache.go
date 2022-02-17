package database

import (
	"database/sql"
	"sync"

	"github.com/heedy/heedy/backend/database/dbutil"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type SqlxCache struct {
	DB                     *sqlx.DB
	Verbose                bool
	preparedStmtCache      map[string]*sqlx.Stmt
	preparedNamedStmtCache map[string]*sqlx.NamedStmt
	lock                   sync.RWMutex
}

// Initializes a sqlx mixin
func (c *SqlxCache) InitCache(sqldb *sqlx.DB) {
	c.DB = sqldb
	c.preparedStmtCache = make(map[string]*sqlx.Stmt)
	c.preparedNamedStmtCache = make(map[string]*sqlx.NamedStmt)
}

// This function returns a prepared statement, or prepares one for the given query
// stores it and returns it
func (db *SqlxCache) GetOrPrepare(query string) (*sqlx.Stmt, error) {
	var err error

	if db.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(query)
	}

	db.lock.RLock()
	prepared, ok := db.preparedStmtCache[query]
	db.lock.RUnlock()
	if ok {
		return prepared, nil
	}

	// Convert to the correct binding type
	prepared, err = db.DB.Preparex(db.DB.Rebind(query))

	if err != nil {
		return prepared, err
	}
	db.lock.Lock()
	db.preparedStmtCache[query] = prepared
	db.lock.Unlock()
	return prepared, nil
}

// This function returns a prepared statement, or prepares one for the given query
// stores it and returns it
func (db *SqlxCache) GetOrPrepareNamed(query string) (*sqlx.NamedStmt, error) {
	var err error

	if db.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(query)
	}

	db.lock.RLock()
	prepared, ok := db.preparedNamedStmtCache[query]
	db.lock.RUnlock()
	if ok {
		return prepared, nil
	}

	// Convert to the correct binding type
	prepared, err = db.DB.PrepareNamed(query)

	if err != nil {
		return prepared, err
	}
	db.lock.Lock()
	db.preparedNamedStmtCache[query] = prepared
	db.lock.Unlock()
	return prepared, nil
}

/**
This is a wrapper for the Get done in sqlx, it does auto conversion to stored
procedures executes them, and does conversion to the given query style for the
given database.

Gets a single item from the DB, remember to add LIMIT 1 if the DB doesn't know
about the query being for a unique item.
**/
func (db *SqlxCache) Get(dest interface{}, query string, args ...interface{}) error {
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
func (db *SqlxCache) Select(dest interface{}, query string, args ...interface{}) error {
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
func (db *SqlxCache) Exec(query string, args ...interface{}) (sql.Result, error) {
	prep, err := db.GetOrPrepare(query)

	if err != nil {
		return nil, err
	}

	return prep.Exec(args...)
}

func (db *SqlxCache) ExecUncached(query string, args ...interface{}) (sql.Result, error) {
	if db.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(query)
	}
	return db.DB.Exec(query, args...)
}

func (db *SqlxCache) NamedExec(query string, arg interface{}) (sql.Result, error) {
	prep, err := db.GetOrPrepareNamed(query)

	if err != nil {
		return nil, err
	}

	return prep.Exec(arg)
}

func (db *SqlxCache) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	prep, err := db.GetOrPrepare(query)

	if err != nil {
		return nil, err
	}

	return prep.Queryx(args...)
}

type TxWrapper struct {
	*sqlx.Tx
	Verbose   bool
	committed bool
}

func (tx TxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	if tx.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(query)
	}
	return tx.Tx.Exec(query, args...)
}
func (tx TxWrapper) Select(dest interface{}, query string, args ...interface{}) error {
	if tx.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(query)
	}
	return tx.Tx.Select(dest, query, args...)
}

func (tx TxWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	if tx.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(query)
	}

	return tx.Tx.Get(dest, query, args...)
}

func (tx TxWrapper) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	if tx.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(query)
	}

	return tx.Tx.Queryx(query, args...)
}

func (tx TxWrapper) Rollback() error {
	if tx.committed {
		return nil
	}
	if tx.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug("ROLLBACK")
	}
	return tx.Tx.Rollback()
}

func (tx *TxWrapper) Commit() error {
	if tx.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug("COMMIT")

	}
	tx.committed = true
	return tx.Tx.Commit()
}

func (db *SqlxCache) Beginx() (TxWrapper, error) {
	if db.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug("BEGIN TRANSACTION")
	}
	tx, err := db.DB.Beginx()
	return TxWrapper{
		Tx:      tx,
		Verbose: db.Verbose,
	}, err
}

func (db *SqlxCache) BeginImmediatex() (TxWrapper, error) {
	if db.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug("BEGIN IMMEDIATE TRANSACTION")
	}
	tx, err := db.DB.Beginx()

	// https://github.com/mattn/go-sqlite3/issues/400
	if err == nil {
		_, err = tx.Exec("ROLLBACK; BEGIN IMMEDIATE")
		if err != nil {
			tx.Rollback()
		}
	}

	return TxWrapper{
		Tx:      tx,
		Verbose: db.Verbose,
	}, err
}
