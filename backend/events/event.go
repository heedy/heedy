package events

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type Event struct {
	Event      string `json:"event"`
	User       string `json:"user,omitempty"`
	Connection string `json:"connection,omitempty"`
	Plugin     string `json:"plugin,omitempty"`
	Source     string `json:"source,omitempty"`
	Key        string `json:"key,omitempty"`
	Type       string `json:"type,omitempty"`

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
	logrus.Debugf("Event: %s", e.String())
	el.Handler.Fire(e)
}

type AsyncFire struct {
	Handler
}

func (af AsyncFire) Fire(e *Event) {
	go af.Handler.Fire(e)
}

// We require a global event manager for sqlite's global hooks
var GlobalManager = EventLogger{NewMultiHandler()}

func Fire(e *Event) {
	GlobalManager.Fire(e)
}

func AddHandler(er Handler) {
	GlobalManager.Handler.(*MultiHandler).AddHandler(er)
}

func RemoveHandler(er Handler) {
	GlobalManager.Handler.(*MultiHandler).RemoveHandler(er)
}
