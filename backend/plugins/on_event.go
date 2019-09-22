package plugins

import (
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/events"

	"github.com/sirupsen/logrus"
)

type EventPostHandler struct {
	Plugin string
	Post   string
}

func (eph *EventPostHandler) Fire(e *events.Event) {
	logrus.Debugf("%s: %s <- %s", eph.Plugin, eph.Post, e.String())
}

func PluginEventHandler(plugin string, e *assets.Event) (*EventPostHandler, error) {
	return &EventPostHandler{
		Plugin: plugin,
		Post:   *e.Post,
	}, nil
}
