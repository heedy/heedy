package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"

	"github.com/heedy/heedy/api/golang/plugin"
	"github.com/heedy/heedy/plugins/registry/backend/api"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info(fmt.Sprintf("%s plugin starting", api.PluginName))
	p, err := plugin.Init()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	pluginMiddleware := plugin.NewMiddleware(p, api.Handler)

	server := http.Server{
		Handler: pluginMiddleware,
	}

	sockPath := fmt.Sprintf("%s.sock", api.PluginName)
	unixListener, err := net.Listen("unix", path.Join(p.Meta.DataDir, sockPath))
	if err != nil {
		p.Logger().Error(fmt.Errorf("Failed to listen on socket: %w", err))
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
	os.Remove(path.Join(p.Meta.DataDir, sockPath))
}
