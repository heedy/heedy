package plugins

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/events"

	"github.com/sirupsen/logrus"
)

var mainClient = http.Client{}

type EventPostHandler struct {
	Plugin string
	Post   string
}

func (eph *EventPostHandler) Fire(e *events.Event) {
	logrus.Debugf("%s: %s <- %s", eph.Plugin, eph.Post, e.String())
	r, err := http.NewRequest("POST", eph.Post, strings.NewReader(e.String()))
	if err != nil {
		logrus.Warn(err)
		return
	}
	res, err := mainClient.Do(r)
	if err != nil {
		logrus.Warn(err)
		return
	}
	res.Body.Close()
}

type EventUnixHandler struct {
	Plugin string
	Post   string
	Path   string
	Client http.Client
}

func (euh *EventUnixHandler) Fire(e *events.Event) {
	logrus.Debugf("%s: %s <- %s", euh.Plugin, euh.Post, e.String())
	r, err := http.NewRequest("POST", "http://unix"+euh.Path, strings.NewReader(e.String()))
	if err != nil {
		logrus.Warn(err)
		return
	}
	res, err := euh.Client.Do(r)
	if err != nil {
		logrus.Warn(err)
		return
	}
	res.Body.Close()
}

func PluginEventHandler(a *assets.Assets, plugin string, e *assets.Event) (events.Handler, error) {
	if e.Post == nil {
		return nil, errors.New("No post was given")
	}
	if !strings.HasPrefix(*e.Post, "unix://") {
		return events.AsyncFire{&EventPostHandler{
			Plugin: plugin,
			Post:   *e.Post,
		}}, nil
	}

	host, path, err := ParseUnixSock(a.DataDir(), *e.Post)
	if err != nil {
		return nil, err
	}
	return events.AsyncFire{&EventUnixHandler{
		Plugin: plugin,
		Post:   *e.Post,
		Path:   path,
		Client: http.Client{
			Transport: &http.Transport{
				DialContext: (&unixDialer{
					Dialer: net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
						DualStack: true,
					},
					Location: host,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}}, nil
}
