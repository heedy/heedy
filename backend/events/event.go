package events

import (
	"encoding/json"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database/dbutil"

	"github.com/sirupsen/logrus"
)

type Event struct {
	Event  string              `json:"event"`
	User   string              `json:"user,omitempty" db:"user"`
	App    string              `json:"app,omitempty" db:"app"`
	Plugin *string             `json:"plugin,omitempty" db:"plugin"`
	Key    *string             `json:"key,omitempty" db:"key"`
	Object string              `json:"object,omitempty" db:"object"`
	Tags   *dbutil.StringArray `json:"tags,omitempty" db:"tags"`
	Type   string              `json:"type,omitempty" db:"type"`

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
		logrus.WithField("stack", dbutil.MiniStack(1)).Debug(e)
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
