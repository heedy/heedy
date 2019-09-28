package events

import (
	"encoding/json"
	"errors"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/assets"

	"github.com/sirupsen/logrus"
)

type Event struct {
	Event      string `json:"event"`
	User       string `json:"user,omitempty" db:"user"`
	Connection string `json:"connection,omitempty" db:"connection"`
	Plugin     string `json:"plugin,omitempty" db:"plugin"`
	Source     string `json:"source,omitempty" db:"source"`
	Key        string `json:"key,omitempty" db:"key"`
	Type       string `json:"type,omitempty" db:"type"`

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
		logrus.WithField("stack",database.MiniStack(1)).Debugf(e.String())
	} else {
		logrus.Debugf(e.String())
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
	if e.Source != "" {
		return db.Get(e, "SELECT sources.owner AS user,sources.connection,connections.plugin,sources.key,sources.type FROM sources LEFT JOIN connections ON sources.connection=connections.id WHERE sources.id=? LIMIT 1", e.Source)
	}
	if e.Connection != "" {
		e.Key = ""
		e.Type = ""
		return db.Get(e, "SELECT owner AS user,plugin FROM connections WHERE id=? LIMIT 1", e.Connection)
	}
	if e.User != "" {
		e.Key = ""
		e.Type = ""
		e.Connection = ""
		e.Plugin = ""
		// This is only to make sure the user exists
		return db.Get(e, "SELECT username AS user FROM users WHERE username=? LIMIT 1", e.User)
	}
	return errors.New("bad_request: An event must target a specific user,connection or source")
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
		logrus.Error(err)
	} else {
		fh.Handler.Fire(e)
	}

}
