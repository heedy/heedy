package plugins

import (
	"errors"
	"net/http"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/sirupsen/logrus"
)

type PluginEventHandler struct {
	Plugin  string
	Post    string
	Handler http.Handler
}

func NewPluginEventHandler(p *Plugin, e *assets.Event) (*PluginEventHandler, error) {
	if e.Post == nil {
		return nil, errors.New("Plugin event doesn't have post")
	}
	h, err := p.Run.GetHandler(p.Name, *e.Post)
	return &PluginEventHandler{
		Plugin:  p.Name,
		Post:    *e.Post,
		Handler: h,
	}, err
}

func (eh *PluginEventHandler) Fire(e *events.Event) {
	logrus.Debugf("%s: %s <- %s", eh.Plugin, eh.Post, e.String())
	_, err := run.Request(eh.Handler, "POST", "/", e, nil)
	if err != nil {
		logrus.Warn(err)
	}
}
