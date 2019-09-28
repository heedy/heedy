package events

import (
	"database/sql"
	"database/sql/driver"
	"github.com/mattn/go-sqlite3"
	"sync"
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/assets"
)

type QueryType int
const (
	SQL_CREATE QueryType= iota
	SQL_UPDATE 
	SQL_DELETE
)

type SqliteHookData struct {
	Type QueryType
	Table string
	RowID int64
	Conn *sqlite3.SQLiteConn
}

type SqliteHook struct {
	Table string	// The table
	Query QueryType	// The op on the table
}

var queryTypeMap = map[int]QueryType{18:SQL_CREATE,23:SQL_UPDATE,9:SQL_DELETE}
var sqliteEventType = make(map[SqliteHook]func(SqliteHookData))

func AddSQLHook(table string, queryType QueryType, hookfunc func(SqliteHookData)) error {
	sqliteEventType[SqliteHook{table,queryType}] = hookfunc
	return nil
}

func connectHook(conn *sqlite3.SQLiteConn) error {
	conn.RegisterUpdateHook(func(op int, dbname string, tblname string, rowid int64) {
		if op == 9 || dbname != "main" {
			return
		}
		qtype,ok := queryTypeMap[op]
		if !ok {
			return
		}
		ename, ok := sqliteEventType[SqliteHook{tblname, qtype}]
		if ok {
			ename(SqliteHookData{
				Type: qtype,
				Table: tblname,
				RowID: rowid,
				Conn: conn,
			})
		}
	})
	conn.RegisterPreUpdateHook(func(pud sqlite3.SQLitePreUpdateData) {
		if pud.Op != 9 || pud.DatabaseName != "main" {
			return
		}

		// We need pre-updates to handle DELETEs, since we need to know the
		// values before they are deleted
		qtype,ok := queryTypeMap[pud.Op]
		if !ok {
			return
		}
		ename, ok := sqliteEventType[SqliteHook{pud.TableName, qtype}]
		if ok {
			ename(SqliteHookData{
				Type: qtype,
				Table: pud.TableName,
				RowID: pud.OldRowID,
				Conn: conn,
			})
		}
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

func SQLiteSelectConn(c *sqlite3.SQLiteConn, stmt string,vals ...driver.Value) (driver.Rows,error) {
	if assets.Get().Config.Verbose {
		logrus.WithField("stack",database.MiniStack(2)).Debug(stmt)
	}
	stmtMutex.RLock()
	s,ok := stmtMap[sqliteConnStmt{c,stmt}]
	stmtMutex.RUnlock()
	if !ok {
		stmtMutex.Lock()
		s,ok = stmtMap[sqliteConnStmt{c,stmt}]
		if !ok {
			var err error
			s,err = c.Prepare(stmt)
			if err!=nil {
				stmtMutex.Unlock()
				return nil,err
			}
			stmtMap[sqliteConnStmt{c,stmt}] = s
		}
		stmtMutex.Unlock()
	}

	return s.Query(vals)

}


// This needs to run before the database is opened, because sqlite3 can only hold a single global
// change listener for each connection, and it must be registered here, rather than on database open, since
// we don't have access to the go-sqlite3 api when opening the database
func init() {
	sql.Register("sqlite3_heedy", &sqlite3.SQLiteDriver{
		ConnectHook: connectHook,
	})
}