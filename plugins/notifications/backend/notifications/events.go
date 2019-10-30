package notifications

import (
	"database/sql/driver"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/events"
)

func getNotification(c *sqlite3.SQLiteConn, stmt string, rowid int64) (*Notification, error) {
	colnum := 10
	rows, err := events.SQLiteSelectConn(c, stmt, rowid)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("Sqlite hook error %w", err)
	}
	vals := make([]driver.Value, colnum)
	for i := 0; i < colnum; i++ {
		var v interface{}
		vals[i] = v

	}
	err = rows.Next(vals)
	if err != nil {
		return nil, fmt.Errorf("Error reading row from sqlite hook %w", err)
	}
	if len(vals) != colnum {
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

	n := &Notification{
		Key:       tsel(vals[0]),
		Timestamp: vals[1].(float64),
	}
	title := tsel(vals[2])
	n.Title = &title
	description := tsel(vals[3])
	n.Description = &description
	ntype := tsel(vals[4])
	n.Type = &ntype
	seen := vals[5].(bool)
	n.Seen = &seen
	user := tsel(vals[6])
	n.User = &user
	global := vals[7].(bool)
	n.Global = &global
	if vals[8] != nil {
		connection := tsel(vals[8])
		n.Connection = &connection
	}
	if vals[9] != nil {
		source := tsel(vals[9])
		n.Source = &source
	}

	return n, nil
}

var notificationEventType = map[events.SqliteHook]string{
	events.SqliteHook{"notifications_user", events.SQL_CREATE}:       "user_notification_create",
	events.SqliteHook{"notifications_connection", events.SQL_CREATE}: "connection_notification_create",
	events.SqliteHook{"notifications_source", events.SQL_CREATE}:     "source_notification_create",
	events.SqliteHook{"notifications_user", events.SQL_UPDATE}:       "user_notification_update",
	events.SqliteHook{"notifications_connection", events.SQL_UPDATE}: "connection_notification_update",
	events.SqliteHook{"notifications_source", events.SQL_UPDATE}:     "source_notification_update",
	events.SqliteHook{"notifications_user", events.SQL_DELETE}:       "user_notification_delete",
	events.SqliteHook{"notifications_connection", events.SQL_DELETE}: "connection_notification_delete",
	events.SqliteHook{"notifications_source", events.SQL_DELETE}:     "source_notification_delete",
}

func RegisterNotificationHooks(e events.Handler) {

	databaseHook := func(s events.SqliteHookData) {
		getStmt := func(tblname string) string {
			switch tblname {
			case "notifications_user":
				return "SELECT key,timestamp,title,description,type,seen,user,global,NULL,NULL FROM notifications_user WHERE rowid=?"
			case "notifications_connection":
				return "SELECT key,timestamp,title,description,type,seen,user,global,connection,NULL FROM notifications_connection WHERE rowid=?"
			case "notifications_source":
				return "SELECT key,timestamp,title,description,type,seen,user,global,connection,source FROM sources LEFT JOIN connections ON sources.connection=connections.id WHERE sources.rowid=?"
			default:
				panic("Unrecognized table name in getStmt")

			}
		}

		n, err := getNotification(s.Conn, getStmt(s.Table), s.RowID)
		if err != nil {
			logrus.Error(err)
			return
		}
		evt := &events.Event{
			Data: n,
		}
		if n.Source != nil {
			evt.Source = *n.Source
		} else if n.Connection != nil {
			evt.Connection = *n.Connection
		} else {
			evt.User = *n.User
		}

		evt.Event = notificationEventType[events.SqliteHook{s.Table, s.Type}]
		go e.Fire(evt)
	}

	events.AddSQLHook("notifications_user", events.SQL_CREATE, databaseHook)
	events.AddSQLHook("notifications_connection", events.SQL_CREATE, databaseHook)
	events.AddSQLHook("notifications_source", events.SQL_CREATE, databaseHook)
	events.AddSQLHook("notifications_user", events.SQL_UPDATE, databaseHook)
	events.AddSQLHook("notifications_connection", events.SQL_UPDATE, databaseHook)
	events.AddSQLHook("notifications_source", events.SQL_UPDATE, databaseHook)
	events.AddSQLHook("notifications_user", events.SQL_DELETE, databaseHook)
	events.AddSQLHook("notifications_connection", events.SQL_DELETE, databaseHook)
	events.AddSQLHook("notifications_source", events.SQL_DELETE, databaseHook)
}
