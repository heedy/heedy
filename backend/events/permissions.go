package events

import (
	"errors"
	"github.com/heedy/heedy/backend/database"
	"strings"
)

var ErrAccessDenied = errors.New("access_denied: You don't have necessary permissions for the given query")

// Check if the given DB can perform the given subscription
func CanSubscribe(db database.DB, e *Event) error {
	dbid := db.ID()
	if dbid == "heedy" {
		return nil
	}
	if e.User == "" && e.App == "" && e.Source == "" {
		return ErrAccessDenied
	}

	i := strings.Index(dbid, "/")
	if i > -1 {
		//usr := dbid[:i]
		conn := dbid[i+1:]
		if e.User != "" || e.App != "" && e.App != conn {
			return ErrAccessDenied
		}
		if e.Source != "" {
			_, err := db.ReadSource(e.Source, nil)
			if err != nil {
				return err
			}
		}
	} else {
		if e.User != "" && e.User != dbid {
			return ErrAccessDenied
		}
		if e.App != "" {
			_, err := db.ReadApp(e.App, nil)
			if err != nil {
				return err
			}
		}
		if e.Source != "" {
			_, err := db.ReadSource(e.Source, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
