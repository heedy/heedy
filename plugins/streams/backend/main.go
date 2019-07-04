package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"

	"github.com/heedy/heedy/backend/plugin"
	"github.com/heedy/heedy/plugins/streams/backend/api"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Streams plugin starting")
	p, err := plugin.Init()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	err = p.InitSQL("streams", api.SQLVersion, api.SQLUpdater)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	pluginMiddleware := plugin.NewMiddleware(p, api.Handler)

	server := http.Server{
		Handler: pluginMiddleware,
	}
	unixListener, err := net.Listen("unix", path.Join(p.Meta.DataDir, "streams.sock"))
	if err != nil {
		p.Logger().Error(err)
		p.Close()
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			server.Close()
		}
	}()

	p.Logger().Info("Plugin Ready")
	server.Serve(unixListener)
	p.Logger().Debug("Closing")
	p.Close()
	os.Remove(path.Join(p.Meta.DataDir, "streams.sock"))

}
