package events

import (
	"encoding/json"
	"errors"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"

	"github.com/sirupsen/logrus"
)

type Event struct {
	Event  string                `json:"event"`
	User   string                `json:"user,omitempty" db:"user"`
	App    string                `json:"app,omitempty" db:"app"`
	Plugin *string               `json:"plugin,omitempty" db:"plugin"`
	Key    *string               `json:"key,omitempty" db:"key"`
	Object string                `json:"object,omitempty" db:"object"`
	Tags   *database.StringArray `json:"tags,omitempty" db:"tags"`
	Type   string                `json:"type,omitempty" db:"type"`

	Data interface{} `json:"data,omitempty"`
}

func (e *Event) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type Handler interface {
	Fire(e *Event)
}

type EventLogger struct {
	Handler
}

func (el EventLogger) Fire(e *Event) {
	if assets.Get().Config.Verbose {
		logrus.WithField("stack", database.MiniStack(1)).Debug(e)
	} else {
		logrus.Debug(e)
	}

	el.Handler.Fire(e)
}

type AsyncFire struct {
	Handler
}

func (af AsyncFire) Fire(e *Event) {
	go af.Handler.Fire(e)
}

// We require a global event manager for sqlite's global hooks
var GlobalHandler = EventLogger{NewMultiHandler()}

func Fire(e *Event) {
	GlobalHandler.Fire(e)
}

func AddHandler(er Handler) {
	GlobalHandler.Handler.(*MultiHandler).AddHandler(er)
}

func RemoveHandler(er Handler) {
	GlobalHandler.Handler.(*MultiHandler).RemoveHandler(er)
}

// FillEvent fills in the event's targeting data
func FillEvent(db *database.AdminDB, e *Event) error {
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
	Handler
	DB *database.AdminDB
}

func NewFilledHandler(db *database.AdminDB, h Handler) FilledHandler {
	return FilledHandler{
		Handler: h,
		DB:      db,
	}
}

func (fh FilledHandler) Fire(e *Event) {
	if err := FillEvent(fh.DB, e); err != nil {
		logrus.Errorf("Failed to validate event %s: %s", e.String(), err)
	} else {
		fh.Handler.Fire(e)
	}

}
