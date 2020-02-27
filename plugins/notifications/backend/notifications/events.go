package notifications

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/events"
)

func getNotification(c *sqlite3.SQLiteConn, stmt string, rowid int64) (*Notification, error) {
	colnum := 12
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
		app := tsel(vals[8])
		n.App = &app
	}
	if vals[9] != nil {
		object := tsel(vals[9])
		n.Object = &object
	}
	if vals[10] != nil {
		aa := make(ActionArray, 0)
		if err := json.Unmarshal([]byte(tsel(vals[10])), &aa); err != nil {
			return nil, err
		}
		n.Actions = &aa
	}
	dismissible := vals[11].(bool)
	n.Dismissible = &dismissible

	return n, nil
}

var notificationEventType = map[events.SqliteHook]string{
	events.SqliteHook{"notifications_user", events.SQL_CREATE}:   "user_notification_create",
	events.SqliteHook{"notifications_app", events.SQL_CREATE}:    "app_notification_create",
	events.SqliteHook{"notifications_object", events.SQL_CREATE}: "object_notification_create",
	events.SqliteHook{"notifications_user", events.SQL_UPDATE}:   "user_notification_update",
	events.SqliteHook{"notifications_app", events.SQL_UPDATE}:    "app_notification_update",
	events.SqliteHook{"notifications_object", events.SQL_UPDATE}: "object_notification_update",
	events.SqliteHook{"notifications_user", events.SQL_DELETE}:   "user_notification_delete",
	events.SqliteHook{"notifications_app", events.SQL_DELETE}:    "app_notification_delete",
	events.SqliteHook{"notifications_object", events.SQL_DELETE}: "object_notification_delete",
}

func RegisterNotificationHooks(e events.Handler) {

	databaseHook := func(s events.SqliteHookData) *events.Event {
		getStmt := func(tblname string) string {
			switch tblname {
			case "notifications_user":
				return "SELECT key,timestamp,title,description,type,seen,user,global,NULL,NULL,actions,dismissible FROM notifications_user WHERE rowid=?"
			case "notifications_app":
				return "SELECT key,timestamp,title,description,type,seen,user,global,app,NULL,actions,dismissible FROM notifications_app WHERE rowid=?"
			case "notifications_object":
				return "SELECT key,timestamp,title,description,type,seen,user,global,app,object,actions,dismissible FROM objects LEFT JOIN apps ON objects.app=apps.id WHERE objects.rowid=?"
			default:
				panic("Unrecognized table name in getStmt")

			}
		}

		n, err := getNotification(s.Conn, getStmt(s.Table), s.RowID)
		if err != nil {
			logrus.Errorf("Failed to process notification event: %s", err)
			return nil
		}
		var evt *events.Event
		if n.Object != nil {
			evt, err = events.FillObjectEvent(s, *n.Object)
			if err != nil {
				logrus.Errorf("Failed to fill object data for notification event: %s", err)
				return nil
			}
		} else if n.App != nil {
			evt, err = events.FillAppEvent(s, *n.App)
			if err != nil {
				logrus.Errorf("Failed to fill app data for notification event: %s", err)
				return nil
			}
		} else {
			evt = &events.Event{
				User: *n.User,
			}
		}
		evt.Data = n
		evt.Event = notificationEventType[events.SqliteHook{s.Table, s.Type}]
		return evt
	}

	events.AddSQLHook("notifications_user", events.SQL_CREATE, databaseHook)
	events.AddSQLHook("notifications_app", events.SQL_CREATE, databaseHook)
	events.AddSQLHook("notifications_object", events.SQL_CREATE, databaseHook)
	events.AddSQLHook("notifications_user", events.SQL_UPDATE, databaseHook)
	events.AddSQLHook("notifications_app", events.SQL_UPDATE, databaseHook)
	events.AddSQLHook("notifications_object", events.SQL_UPDATE, databaseHook)
	events.AddSQLHook("notifications_user", events.SQL_DELETE, databaseHook)
	events.AddSQLHook("notifications_app", events.SQL_DELETE, databaseHook)
	events.AddSQLHook("notifications_object", events.SQL_DELETE, databaseHook)
}
