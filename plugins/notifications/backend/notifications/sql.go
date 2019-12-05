package notifications

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
)

const SQLVersion = 1

const sqlSchema = `
	-- We split up the schema into 3 tables due to issues with UNIQUE when certain values are NULL.
	-- We need apps/objects to be nullable to represent notifications for users/apps
	-- https://stackoverflow.com/questions/22699409/sqlite-null-and-unique

	CREATE TABLE notifications_user (
		user VARCHAR NOT NULL,
		key VARCHAR NOT NULL,

		title VARCHAR NOT NULL,
		description VARCHAR NOT NULL DEFAULT '',
		type VARCHAR NOT NULL DEFAULT 'info',
		timestamp REAL NOT NULL,
		actions VARCHAR NOT NULL DEFAULT '[]',

		-- User notifications are global=true
		global BOOLEAN NOT NULL DEFAULT true,
		dismissible BOOLEAN NOT NULL DEFAULT true,
		seen BOOLEAN NOT NULL DEFAULT false,

		CONSTRAINT pk PRIMARY KEY (user,key),
		CONSTRAINT valid_actions CHECK(json_valid(actions) AND json_type(actions)=='array'),

		CONSTRAINT user_c
			FOREIGN KEY (user)
			REFERENCES users(username)
			ON UPDATE CASCADE
			ON DELETE CASCADE
	);

	CREATE TABLE notifications_app (
		user VARCHAR NOT NULL,
		app VARCHAR NOT NULL,
		key VARCHAR NOT NULL,

		title VARCHAR NOT NULL,
		description VARCHAR NOT NULL DEFAULT '',
		type VARCHAR NOT NULL DEFAULT 'info',
		timestamp REAL NOT NULL,
		actions VARCHAR NOT NULL DEFAULT '[]',

		global BOOLEAN NOT NULL DEFAULT false,
		seen BOOLEAN NOT NULL DEFAULT false,
		dismissible BOOLEAN NOT NULL DEFAULT true,

		CONSTRAINT pk PRIMARY KEY (user,app,key),
		CONSTRAINT valid_actions CHECK(json_valid(actions) AND json_type(actions)=='array'),

		CONSTRAINT user_c
			FOREIGN KEY (user)
			REFERENCES users(username)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT app_c
			FOREIGN KEY (app)
			REFERENCES apps(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE
	);

	CREATE TABLE notifications_object (
		user VARCHAR NOT NULL,
		app VARCHAR NOT NULL,
		object VARCHAR NOT NULL,
		key VARCHAR NOT NULL,

		title VARCHAR NOT NULL,
		description VARCHAR NOT NULL DEFAULT '',
		type VARCHAR NOT NULL DEFAULT 'info',
		actions VARCHAR NOT NULL DEFAULT '[]',
		timestamp REAL NOT NULL,

		global BOOLEAN NOT NULL DEFAULT false,
		seen BOOLEAN NOT NULL DEFAULT false,
		dismissible BOOLEAN NOT NULL DEFAULT true,

		CONSTRAINT pk PRIMARY KEY (user,app,object,key),
		CONSTRAINT valid_actions CHECK(json_valid(actions) AND json_type(actions)=='array'),

		CONSTRAINT user_c
			FOREIGN KEY (user)
			REFERENCES users(username)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT app_c
			FOREIGN KEY (app)
			REFERENCES apps(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT object_c
			FOREIGN KEY (object)
			REFERENCES objects(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE
	);
`

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, i *run.Info, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion != 0 {
		return errors.New("Notifications database version too new")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}

var ErrAccessDenied = errors.New("access_denied: You don't have necessary permissions for the given query")

type Action struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Href        string `json:"href"`
	NewWindow   bool   `json:"new_window"`
	Dismiss     bool   `json:"dismiss"`
}

type ActionArray []Action

func (aa *ActionArray) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, aa)
		return nil
	case string:
		json.Unmarshal([]byte(v), aa)
		return nil
	default:
		return fmt.Errorf("Can't scan json array array, unsupported type: %T", v)
	}
}

func (aa *ActionArray) Value() (driver.Value, error) {
	return json.Marshal(aa)
}

type Notification struct {
	Key       string  `json:"key,omitempty"`
	Timestamp float64 `json:"timestamp,omitempty"`

	User   *string `json:"user,omitempty"`
	App    *string `json:"app,omitempty"`
	Object *string `json:"object,omitempty"`

	Type        *string      `json:"type,omitempty"`
	Title       *string      `json:"title,omitempty"`
	Description *string      `json:"description,omitempty"`
	Actions     *ActionArray `json:"actions,omitempty"`

	Dismissible *bool `json:"dismissible,omitempty"`
	Seen        *bool `json:"seen,omitempty"`
	Global      *bool `json:"global,omitempty"`
}

type NotificationsQuery struct {
	User   *string `json:"user,omitempty" schema:"user"`
	App    *string `json:"app,omitempty" schema:"app"`
	Object *string `json:"object,omitempty" schema:"object"`

	Global      *bool   `json:"global,omitempty" schema:"global"`
	Seen        *bool   `json:"seen,omitempty" schema:"seen"`
	Key         *string `json:"key,omitempty" schema:"key"`
	Dismissible *bool   `json:"dismissible,omitempty" schema:"dismissible"`

	Type *string `json:"type,omitempty"`

	// Whether  or not to include self when * present. For example {user="test",app="*"}
	// is unclear whether the user's notifications should be included or not. False by default
	IncludeSelf *bool `json:"include_self,omitempty" schema:"include_self"`
}

func queryAllowed(db database.DB, o *NotificationsQuery) (*NotificationsQuery, error) {
	if o == nil {
		o = &NotificationsQuery{}
	}
	if o.IncludeSelf == nil {
		includeSelfDefault := false
		o.IncludeSelf = &includeSelfDefault
	}

	dbid := db.ID()
	if dbid == "public" {
		return nil, ErrAccessDenied
	}

	// Set up the query's permissions
	if dbid != "heedy" {
		i := strings.Index(dbid, "/")
		if i > -1 {
			usr := dbid[:i]
			conn := dbid[i+1:]
			if o.User != nil || o.App != nil && *o.App != conn {
				return nil, ErrAccessDenied
			}
			o.User = &usr
			o.App = &conn
		} else {
			if o.User != nil && *o.User != dbid {
				return nil, ErrAccessDenied
			}
			o.User = &dbid
		}
	}
	return o, nil
}

func extractQueryBasics(o *NotificationsQuery) ([]string, []interface{}) {
	cNames := []string{}
	cValues := []interface{}{}
	if o.Type != nil {
		cNames = append(cNames, "type")
		cValues = append(cValues, *o.Type)
	}
	if o.Seen != nil {
		cNames = append(cNames, "seen")
		cValues = append(cValues, *o.Seen)
	}
	if o.Global != nil {
		cNames = append(cNames, "global")
		cValues = append(cValues, *o.Global)
	}
	if o.Key != nil {
		cNames = append(cNames, "key")
		cValues = append(cValues, *o.Key)
	}
	if o.Dismissible != nil {
		cNames = append(cNames, "dismissible")
		cValues = append(cValues, *o.Dismissible)
	}
	return cNames, cValues
}

func includeTable(o *NotificationsQuery) (bool, bool, bool) {
	if o == nil {
		o = &NotificationsQuery{}
	}
	if o.IncludeSelf == nil {
		includeSelfDefault := false
		o.IncludeSelf = &includeSelfDefault
	}

	includeUser := false
	includeApp := false
	includeObject := false
	if o.User == nil && o.App == nil && o.Object == nil {
		includeUser = true
		includeApp = true
		includeObject = true
	} else {
		if o.Object != nil {
			includeObject = true
		}
		if o.App != nil && (*o.App == "*" || o.Object == nil || *o.IncludeSelf || *o.Object != "*") {
			includeApp = true
		}
		if o.User != nil && (*o.User == "*" || (o.Object == nil && o.App == nil) || *o.IncludeSelf || o.Object != nil && *o.Object != "*" || o.App != nil && *o.App != "*") {
			includeUser = true
		}
	}
	return includeUser, includeApp, includeObject
}

// ReadNotifications reads the notifications associated with the given user/app/object
func ReadNotifications(db database.DB, o *NotificationsQuery) ([]Notification, error) {
	// Figure out which tables to query for the results
	includeUser, includeApp, includeObject := includeTable(o)

	o, err := queryAllowed(db, o)
	if err != nil {
		return nil, err
	}

	res := []Notification{}

	// Set up the query that will be used to filter results
	cNames, cValues := extractQueryBasics(o)

	if o.User != nil && *o.User != "*" {
		cNames = append(cNames, "user")
		cValues = append(cValues, *o.User)
	}

	if includeUser {
		queryWhere := strings.Join(cNames, "=? AND ") + "=?"
		var r []Notification
		err := db.AdminDB().Select(&r, fmt.Sprintf("SELECT * FROM notifications_user WHERE %s;", queryWhere), cValues...)
		if err != nil {
			return nil, err
		}
		res = append(res, r...)
	}

	if o.App != nil && *o.App != "*" {
		cNames = append(cNames, "app")
		cValues = append(cValues, *o.App)
	}

	if includeApp {
		queryWhere := strings.Join(cNames, "=? AND ") + "=?"
		var r []Notification
		err := db.AdminDB().Select(&r, fmt.Sprintf("SELECT * FROM notifications_app WHERE %s;", queryWhere), cValues...)
		if err != nil {
			return nil, err
		}
		res = append(res, r...)
	}

	if o.Object != nil && *o.Object != "*" {
		cNames = append(cNames, "object")
		cValues = append(cValues, *o.Object)
	}

	if includeObject {
		queryWhere := strings.Join(cNames, "=? AND ") + "=?"
		var r []Notification
		err := db.AdminDB().Select(&r, fmt.Sprintf("SELECT * FROM notifications_object WHERE %s;", queryWhere), cValues...)
		if err != nil {
			return nil, err
		}
		res = append(res, r...)
	}

	return res, nil
}

func extractNotificationBasics(n *Notification) ([]string, []interface{}) {
	cNames := []string{}
	cValues := []interface{}{}

	if n.Key != "" {
		cNames = append(cNames, "key")
		cValues = append(cValues, n.Key)
	}

	if n.Title != nil {
		cNames = append(cNames, "title")
		cValues = append(cValues, n.Title)
	}
	if n.Description != nil {
		cNames = append(cNames, "description")
		cValues = append(cValues, *n.Description)
	}
	if n.Seen != nil {
		cNames = append(cNames, "seen")
		cValues = append(cValues, *n.Seen)
	}
	if n.Global != nil {
		cNames = append(cNames, "global")
		cValues = append(cValues, *n.Global)
	}
	if n.Type != nil {
		cNames = append(cNames, "type")
		cValues = append(cValues, *n.Type)
	}
	if n.Actions != nil {
		cNames = append(cNames, "actions")
		cValues = append(cValues, n.Actions)
	}
	if n.Dismissible != nil {
		cNames = append(cNames, "dismissible")
		cValues = append(cValues, *n.Dismissible)
	}
	return cNames, cValues
}

func excludeStmt(cNames []string) string {
	narr := make([]string, len(cNames))
	for i := range cNames {
		narr[i] = fmt.Sprintf("%s=excluded.%s", cNames[i], cNames[i])
	}
	return strings.Join(narr, ", ")
}

// WriteNotification writes the given notification. If a notification with the given key and target exists, it updates the existing notification with the new
// values. The notification will only update those values that are specifically set in the new notification
func WriteNotification(db database.DB, n *Notification) error {
	dbid := db.ID()
	if n.Key == "" || n.Title == nil || *n.Title == "" {
		return errors.New("bad_request: Notifications must have a valid key and title")
	}
	if n.Timestamp != 0 {
		return errors.New("bad_request: timestamps are set automatically")
	}
	if n.User == nil && n.App == nil && n.Object == nil && dbid != "heedy" {
		// The notification is to be inserted to itself
		i := strings.Index(dbid, "/")
		if i > -1 {
			conn := dbid[i+1:]
			n.App = &conn
		} else {
			n.User = &dbid
		}
	}

	// Set up the columns that will be set on the notification
	cNames, cValues := extractNotificationBasics(n)
	cNames = append(cNames, "timestamp")
	cValues = append(cValues, float64(time.Now().UnixNano())*1e-9)
	eS := excludeStmt(cNames)

	if n.Object != nil {
		// The notification is for a object
		s, err := db.ReadObject(*n.Object, nil)
		if err != nil {
			return err
		}
		if dbid == "heedy" || dbid == *s.Owner && s.App == nil || s.App != nil && dbid == *s.Owner+"/"+*s.App {
			// Allow writing the notification
			n.User = s.Owner
			n.App = s.App
			cNames = append(cNames, "user", "app", "object")
			cValues = append(cValues, *s.Owner, *s.App, s.ID)
			_, err := db.AdminDB().Exec(fmt.Sprintf("INSERT INTO notifications_object(%s) VALUES (%s) ON CONFLICT(user,key) DO UPDATE SET %s;", strings.Join(cNames, ","), database.QQ(len(cNames)), eS),
				cValues...)
			return err
		}
		return database.ErrAccessDenied("Can't set notifications for this object")
	}
	if n.App != nil {
		c, err := db.ReadApp(*n.App, nil)
		if err != nil {
			return err
		}

		if dbid == "heedy" || dbid == *c.Owner+"/"+c.ID {
			// Allow writing the notification
			n.User = c.Owner
			n.App = &c.ID
			cNames = append(cNames, "user", "app")
			cValues = append(cValues, *c.Owner, c.ID)
			_, err := db.AdminDB().Exec(fmt.Sprintf("INSERT INTO notifications_app(%s) VALUES (%s) ON CONFLICT(user,app,key) DO UPDATE SET %s;", strings.Join(cNames, ","), database.QQ(len(cNames)), eS),
				cValues...)
			return err
		}
		return database.ErrAccessDenied("Can't set notifications for this app")
	}
	if n.User == nil {
		return errors.New("Must specify a target for the notification")
	}
	u, err := db.ReadUser(*n.User, nil)
	if err != nil {
		return err
	}
	if dbid == "heedy" || *u.UserName == dbid {
		cNames = append(cNames, "user")
		cValues = append(cValues, *u.UserName)
		_, err := db.AdminDB().Exec(fmt.Sprintf("INSERT INTO notifications_user(%s) VALUES (%s) ON CONFLICT(user,app,object,key) DO UPDATE SET %s;", strings.Join(cNames, ","), database.QQ(len(cNames)), eS),
			cValues...)
		return err
	}
	return database.ErrAccessDenied("Can't set notifications for this user")
}

// UpdateNotification is a special version that modifies all notifications satisfying the constraints given in NotificationsQuery
func UpdateNotification(db database.DB, n *Notification, o *NotificationsQuery) error {
	includeUser, includeApp, includeObject := includeTable(o)
	if n.Timestamp != 0 {
		return errors.New("bad_request: timestamps are set automatically")
	}

	o, err := queryAllowed(db, o)
	if err != nil {
		return err
	}

	ncNames, ncValues := extractNotificationBasics(n)
	ocNames, ocValues := extractQueryBasics(o)

	queryUpdate := strings.Join(ncNames, "=?, ") + "=?"

	if o.User != nil && *o.User != "*" {
		ocNames = append(ocNames, "user")
		ocValues = append(ocValues, *o.User)
	}

	if includeUser {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("UPDATE notifications_user SET %s WHERE %s", queryUpdate, queryWhere)
		vals := append(append([]interface{}{}, ncValues...), ocValues...)
		_, err := db.AdminDB().Exec(qstring, vals...)
		if err != nil {
			return err
		}
	}

	if o.App != nil && *o.App != "*" {
		ocNames = append(ocNames, "app")
		ocValues = append(ocValues, *o.App)
	}

	if includeApp {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("UPDATE notifications_app SET %s WHERE %s", queryUpdate, queryWhere)
		vals := append(append([]interface{}{}, ncValues...), ocValues...)
		_, err := db.AdminDB().Exec(qstring, vals...)
		if err != nil {
			return err
		}
	}

	if o.Object != nil && *o.Object != "*" {
		ocNames = append(ocNames, "object")
		ocValues = append(ocValues, *o.Object)
	}

	if includeObject {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("UPDATE notifications_object SET %s WHERE %s", queryUpdate, queryWhere)
		vals := append(append([]interface{}{}, ncValues...), ocValues...)
		_, err := db.AdminDB().Exec(qstring, vals...)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteNotification takes a queryer for notifications
func DeleteNotification(db database.DB, o *NotificationsQuery) error {
	includeUser, includeApp, includeObject := includeTable(o)
	o, err := queryAllowed(db, o)
	if err != nil {
		return err
	}

	ocNames, ocValues := extractQueryBasics(o)

	if o.User != nil && *o.User != "*" {
		ocNames = append(ocNames, "user")
		ocValues = append(ocValues, *o.User)
	}

	if includeUser {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("DELETE FROM notifications_user WHERE %s", queryWhere)
		_, err := db.AdminDB().Exec(qstring, ocValues...)
		if err != nil {
			return err
		}
	}

	if o.App != nil && *o.App != "*" {
		ocNames = append(ocNames, "app")
		ocValues = append(ocValues, *o.App)
	}

	if includeApp {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("DELETE FROM notifications_app WHERE %s", queryWhere)
		_, err := db.AdminDB().Exec(qstring, ocValues...)
		if err != nil {
			return err
		}
	}

	if o.Object != nil && *o.Object != "*" {
		ocNames = append(ocNames, "object")
		ocValues = append(ocValues, *o.Object)
	}

	if includeObject {
		queryWhere := strings.Join(ocNames, "=? AND ") + "=?"
		qstring := fmt.Sprintf("DELETE FROM notifications_object WHERE %s", queryWhere)
		_, err := db.AdminDB().Exec(qstring, ocValues...)
		if err != nil {
			return err
		}
	}

	return nil

}
