package events

import (
	"container/list"
	"database/sql"
	"database/sql/driver"
	"sync"

	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database/dbutil"
)

type QueryType int

const (
	SQL_CREATE QueryType = iota
	SQL_UPDATE
	SQL_DELETE
)

type SqliteHookData struct {
	Type  QueryType
	Table string
	RowID int64
	Conn  *sqlite3.SQLiteConn
}

type SqliteHook struct {
	Table string    // The table
	Query QueryType // The op on the table
}

var queryTypeMap = map[int]QueryType{18: SQL_CREATE, 23: SQL_UPDATE, 9: SQL_DELETE}
var sqliteEventType = make(map[SqliteHook]func(SqliteHookData) *Event)

func AddSQLHook(table string, queryType QueryType, hookfunc func(SqliteHookData) *Event) error {
	sqliteEventType[SqliteHook{table, queryType}] = hookfunc
	return nil
}

func connectHook(conn *sqlite3.SQLiteConn) error {
	// We keep a list of events that we are processing, before the database undergoes a commit
	elist := list.New()
	conn.RegisterUpdateHook(func(op int, dbname string, tblname string, rowid int64) {
		if op == 9 || dbname != "main" {
			return
		}
		qtype, ok := queryTypeMap[op]
		if !ok {
			return
		}
		ename, ok := sqliteEventType[SqliteHook{tblname, qtype}]
		if ok {
			evt := ename(SqliteHookData{
				Type:  qtype,
				Table: tblname,
				RowID: rowid,
				Conn:  conn,
			})
			if evt != nil {
				if assets.Get().Config.Verbose {
					logrus.WithField("stack", dbutil.MiniStack(2)).Debugf("Preparing event %s", evt.String())
				}
				elist.PushBack(evt)
			}
		}
	})
	conn.RegisterPreUpdateHook(func(pud sqlite3.SQLitePreUpdateData) {
		if pud.Op != 9 || pud.DatabaseName != "main" {
			return
		}

		// We need pre-updates to handle DELETEs, since we need to know the
		// values before they are deleted
		qtype, ok := queryTypeMap[pud.Op]
		if !ok {
			return
		}
		ename, ok := sqliteEventType[SqliteHook{pud.TableName, qtype}]
		if ok {
			evt := ename(SqliteHookData{
				Type:  qtype,
				Table: pud.TableName,
				RowID: pud.OldRowID,
				Conn:  conn,
			})
			if evt != nil {
				if assets.Get().Config.Verbose {
					logrus.WithField("stack", dbutil.MiniStack(2)).Debugf("Preparing event %s", evt.String())
				}
				elist.PushBack(evt)
			}
		}
	})

	// The above hooks are called on each update to the database, before they are committed.
	// We actually want to fire the events only AFTER they are committed, so that transaction rollbacks don't mess with us.

	conn.RegisterCommitHook(func() int {
		el2 := elist
		elist = list.New()

		// Want to let the event firing to happen asynchronously, since we want the commit to finish ASAP
		go func() {
			// The transaction was committed, so fire the events
			if assets.Get().Config.Verbose {
				ll := el2.Len()
				if ll > 0 {
					logrus.Debugf("Database commit - firing %d prepared event(s)", ll)
				}
			}
			el := el2.Front()
			for el != nil {
				go Fire(el.Value.(*Event))
				el = el.Next()
			}
		}()

		return 0
	})
	conn.RegisterRollbackHook(func() {
		// There was a rollback, so we get rid of the cached events
		ll := elist.Len()
		if ll > 0 {
			logrus.Debugf("Database rollback detected, cancelling %d prepared event(s)", ll)
		}
		elist.Init()
	})
	return nil
}

// Select on the raw conn, with prepared statements
type sqliteConnStmt struct {
	Conn *sqlite3.SQLiteConn
	Stmt string
}

var stmtMutex = sync.RWMutex{}
var stmtMap = make(map[sqliteConnStmt]driver.Stmt)

func SQLiteSelectConn(c *sqlite3.SQLiteConn, stmt string, vals ...driver.Value) (driver.Rows, error) {
	if assets.Get().Config.Verbose {
		logrus.WithField("stack", dbutil.MiniStack(2)).Debug(stmt)
	}
	stmtMutex.RLock()
	s, ok := stmtMap[sqliteConnStmt{c, stmt}]
	stmtMutex.RUnlock()
	if !ok {
		stmtMutex.Lock()
		s, ok = stmtMap[sqliteConnStmt{c, stmt}]
		if !ok {
			var err error
			s, err = c.Prepare(stmt)
			if err != nil {
				stmtMutex.Unlock()
				return nil, err
			}
			stmtMap[sqliteConnStmt{c, stmt}] = s
		}
		stmtMutex.Unlock()
	}

	return s.Query(vals)

}

// This needs to run before the database is opened, because sqlite3 can only hold a single global
// change listener for each app, and it must be registered here, rather than on database open, since
// we don't have access to the go-sqlite3 api when opening the database
func init() {
	sql.Register("sqlite3_heedy", &sqlite3.SQLiteDriver{
		ConnectHook: connectHook,
	})
}
