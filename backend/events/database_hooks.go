package events

import (
	"database/sql/driver"

	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var databaseEventType = map[SqliteHook]string{
	SqliteHook{"users", SQL_CREATE}:       "user_create",
	SqliteHook{"apps", SQL_CREATE}: "app_create",
	SqliteHook{"objects", SQL_CREATE}:     "object_create",
	SqliteHook{"users", SQL_UPDATE}:       "user_update",
	SqliteHook{"apps", SQL_UPDATE}: "app_update",
	SqliteHook{"objects", SQL_UPDATE}:     "object_update",
	SqliteHook{"users", SQL_DELETE}:       "user_delete",
	SqliteHook{"apps", SQL_DELETE}: "app_delete",
	SqliteHook{"objects", SQL_DELETE}:     "object_delete",
}

// getEvent returns the username, app id, and object id associated with the given event.
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
		App: tsel(vals[1]),
		Plugin:     &plugin,
		Object:     tsel(vals[3]),
		Key:        tsel(vals[4]),
		Type:       tsel(vals[5]),
	}, nil
}

func databaseHook(s SqliteHookData) {
	getStmt := func(tblname string) string {
		switch tblname {
		case "users":
			return "SELECT username,'','','','','' FROM users WHERE rowid=?"
		case "apps":
			return "SELECT owner,id,plugin,'','','' FROM apps WHERE rowid=?"
		case "objects":
			return "SELECT objects.owner,objects.app,apps.plugin,objects.id,objects.key,objects.type FROM objects LEFT JOIN apps ON objects.app=apps.id WHERE objects.rowid=?"
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
	AddSQLHook("apps", SQL_CREATE, databaseHook)
	AddSQLHook("objects", SQL_CREATE, databaseHook)
	AddSQLHook("users", SQL_UPDATE, databaseHook)
	AddSQLHook("apps", SQL_UPDATE, databaseHook)
	AddSQLHook("objects", SQL_UPDATE, databaseHook)
	AddSQLHook("users", SQL_DELETE, databaseHook)
	AddSQLHook("apps", SQL_DELETE, databaseHook)
	AddSQLHook("objects", SQL_DELETE, databaseHook)
}
