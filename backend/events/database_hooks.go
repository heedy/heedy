package events

import (
	"database/sql/driver"

	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var databaseEventType = map[SqliteHook]string{
	SqliteHook{"users", SQL_CREATE}:       "user_create",
	SqliteHook{"connections", SQL_CREATE}: "connection_create",
	SqliteHook{"sources", SQL_CREATE}:     "source_create",
	SqliteHook{"users", SQL_UPDATE}:       "user_update",
	SqliteHook{"connections", SQL_UPDATE}: "connection_update",
	SqliteHook{"sources", SQL_UPDATE}:     "source_update",
	SqliteHook{"users", SQL_DELETE}:       "user_delete",
	SqliteHook{"connections", SQL_DELETE}: "connection_delete",
	SqliteHook{"sources", SQL_DELETE}:     "source_delete",
}

// getEvent returns the username, connection id, and source id associated with the given event.
// The associated stmt should automatically return empty strings for inapplicable values
func getEvent(c *sqlite3.SQLiteConn, stmt string, rowid int64) (*Event, error) {
	rows, err := SQLiteSelectConn(c, stmt, rowid)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("Sqlite hook error %w", err)
	}
	vals := make([]driver.Value, 6)
	for i := 0; i < 6; i++ {
		var v interface{}
		vals[i] = v

	}
	err = rows.Next(vals)
	if err != nil {
		return nil, fmt.Errorf("Error reading row from sqlite hook %w", err)
	}
	if len(vals) != 6 {
		return nil, fmt.Errorf("Sqlite hook: Incorrect number of returned results")
	}

	tsel := func(v interface{}) string {
		switch vv := v.(type) {
		case string:
			return vv
		case []byte:
			return string(vv)
		default:
			return ""
		}
	}

	plugin := tsel(vals[2])

	return &Event{
		User:       tsel(vals[0]),
		Connection: tsel(vals[1]),
		Plugin:     &plugin,
		Source:     tsel(vals[3]),
		Key:        tsel(vals[4]),
		Type:       tsel(vals[5]),
	}, nil
}

func databaseHook(s SqliteHookData) {
	getStmt := func(tblname string) string {
		switch tblname {
		case "users":
			return "SELECT username,'','','','','' FROM users WHERE rowid=?"
		case "connections":
			return "SELECT owner,id,plugin,'','','' FROM connections WHERE rowid=?"
		case "sources":
			return "SELECT sources.owner,sources.connection,connections.plugin,sources.id,sources.key,sources.type FROM sources LEFT JOIN connections ON sources.connection=connections.id WHERE sources.rowid=?"
		default:
			panic("Unrecognized table name in getStmt")

		}
	}

	evt, err := getEvent(s.Conn, getStmt(s.Table), s.RowID)
	if err != nil {
		logrus.Error(err)
		return

	}
	evt.Event = databaseEventType[SqliteHook{s.Table, s.Type}]
	go Fire(evt)

}

func RegisterDatabaseHooks() {
	AddSQLHook("users", SQL_CREATE, databaseHook)
	AddSQLHook("connections", SQL_CREATE, databaseHook)
	AddSQLHook("sources", SQL_CREATE, databaseHook)
	AddSQLHook("users", SQL_UPDATE, databaseHook)
	AddSQLHook("connections", SQL_UPDATE, databaseHook)
	AddSQLHook("sources", SQL_UPDATE, databaseHook)
	AddSQLHook("users", SQL_DELETE, databaseHook)
	AddSQLHook("connections", SQL_DELETE, databaseHook)
	AddSQLHook("sources", SQL_DELETE, databaseHook)
}
