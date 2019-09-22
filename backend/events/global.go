package events

import (
	"github.com/sirupsen/logrus"
)

type EventLogger struct {
	Handler
}

func (el EventLogger) Fire(e *Event) {
	logrus.Debugf("Event: %s", e.String())
	el.Handler.Fire(e)
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
