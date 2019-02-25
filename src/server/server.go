package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/connectordb/connectordb/plugin"
	"github.com/spf13/afero"

	"github.com/connectordb/connectordb/assets"

	log "github.com/sirupsen/logrus"
)

func Run(a *assets.Assets) error {

	ph, err := plugin.NewManager(a)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Info("Cleanup...")
			d, _ := time.ParseDuration("5s")
			ph.Stop(d)
			log.Info("Done")
			os.Exit(0)
		}
	}()

	serverAddress := fmt.Sprintf("%s:%d", *a.Config.Host, *a.Config.Port)

	restMux, err := RestMux()
	if err != nil {
		log.Panic(err)
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/cdb/", restMux)
	mux.Handle("/", http.FileServer(afero.NewHttpFs(a.FS)))

	handler := http.Handler(mux)

	if ph.Middleware != nil {
		log.Info("Adding plugin middleware")
		handler = ph.Middleware(handler)
	}

	// the grpcHandlerFunc takes an grpc server and a http muxer and will
	// route the request to the right place at runtime.
	mergeHandler := grpcHandlerFunc(grpcServer, handler)

	// configure TLS for our server. TLS is REQUIRED to make this setup work.
	// check https://golang.org/src/net/http/server.go?#L2746
	if err != nil {
		log.Panic(err)
	}

	srv := &http.Server{
		Addr:    serverAddress,
		Handler: mergeHandler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*crt},
			NextProtos:   []string{"h2"},
			//InsecureSkipVerify: true,
		},
	}

	// Set up a http listener
	if *a.Config.HTTPPort > 0 {
		httpServer := fmt.Sprintf("%s:%d", *a.Config.Host, *a.Config.HTTPPort)
		log.Infof("Starting http server at %s", httpServer)
		go http.ListenAndServe(httpServer, handler)
	}
	// start listening on the socket
	// Note that if you listen on localhost:<port> you'll not be able to accept
	// connections over the network. Change it to ":port"  if you want it.
	conn, err := net.Listen("tcp", serverAddress)
	if err != nil {
		return err
	}

	// start the server
	log.Infof("starting GRPC and REST on %s", serverAddress)
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return err

}
