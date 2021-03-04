package database

import (
	"errors"
	"strings"

	"github.com/heedy/heedy/backend/events"
	"github.com/sirupsen/logrus"
)

// FillEvent fills in the event's targeting data
func FillEvent(db *AdminDB, e *events.Event) error {
	if e.Event == "" {
		return errors.New("bad_request: No event type specified")
	}
	if e.Object != "" {
		return db.Get(e, "SELECT objects.owner AS user,COALESCE(objects.app,'') AS app,apps.plugin,objects.tags AS tags,objects.key AS key,objects.type FROM objects LEFT JOIN apps ON objects.app=apps.id WHERE objects.id=? LIMIT 1", e.Object)
	}
	if e.App != "" {
		e.Tags = nil
		e.Key = nil
		e.Type = ""
		e.Plugin = nil
		return db.Get(e, "SELECT owner AS user,plugin FROM apps WHERE id=? LIMIT 1", e.App)
	}
	if e.User != "" {
		e.Tags = nil
		e.Type = ""
		e.App = ""
		e.Key = nil
		e.Plugin = nil
		// This is only to make sure the user exists
		return db.Get(e, "SELECT username AS user FROM users WHERE username=? LIMIT 1", e.User)
	}
	return errors.New("bad_request: An event must target a specific user,app or object")
}

type FilledHandler struct {
	events.Handler
	DB *AdminDB
}

func NewFilledHandler(db *AdminDB, h events.Handler) FilledHandler {
	return FilledHandler{
		Handler: h,
		DB:      db,
	}
}

func (fh FilledHandler) Fire(e *events.Event) {
	if err := FillEvent(fh.DB, e); err != nil {
		logrus.Errorf("Failed to validate event %s: %s", e.String(), err)
	} else {
		fh.Handler.Fire(e)
	}

}

var ErrSAccessDenied = errors.New("access_denied: You don't have necessary permissions for the given query")

// Check if the given DB can perform the given subscription
func CanSubscribe(db DB, e *events.Event) error {
	dbid := db.ID()
	if dbid == "heedy" {
		return nil
	}
	if e.User == "" && e.App == "" && e.Object == "" {
		return ErrSAccessDenied
	}

	i := strings.Index(dbid, "/")
	if i > -1 {
		//usr := dbid[:i]
		conn := dbid[i+1:]
		if e.User != "" || e.App != "" && e.App != conn {
			return ErrSAccessDenied
		}
		if e.Object != "" {
			_, err := db.ReadObject(e.Object, nil)
			if err != nil {
				return err
			}
		}
	} else {
		if e.User != "" && e.User != dbid {
			return ErrSAccessDenied
		}
		if e.App != "" {
			_, err := db.ReadApp(e.App, nil)
			if err != nil {
				return err
			}
		}
		if e.Object != "" {
			_, err := db.ReadObject(e.Object, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
