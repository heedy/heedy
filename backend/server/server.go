package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var queryDecoder = schema.NewDecoder()

// RunOptions give special options for running
type RunOptions struct {
	Verbose bool
}

func Run(r *RunOptions) error {
	db, err := database.Open(assets.Get())
	if err != nil {
		return err
	}

	auth := NewAuth(db)

	serverAddress := fmt.Sprintf("%s:%d", assets.Config().GetHost(), assets.Config().GetPort())

	apiMux, err := APIMux()
	if err != nil {
		return err
	}
	authMux, err := AuthMux(auth)
	if err != nil {
		return err
	}
	fMux, err := FrontendMux()
	if err != nil {
		return err
	}

	mux := chi.NewMux()
	mux.Mount("/api", apiMux)
	mux.Mount("/auth", authMux)
	mux.Mount("/", fMux)

	// Get assets directly for the main files

	//mux.Handle("/app", assetFS)
	//mux.Handle("/www/*", assetFS)
	/*
		mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/app/", http.StatusFound)
		})
	*/
	sm, err := NewSourceManager(assets.Get(), http.Handler(mux))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	om, err := NewOverlayManager(assets.Get(), sm)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	em := NewExecManager(assets.Get())
	err = em.Start()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Info("Cleanup...")
			em.Stop()
			db.Close()
			log.Info("Done")
			os.Exit(0)
		}
	}()

	requestHandler := http.Handler(NewRequestHandler(auth, em, om))

	if r != nil && r.Verbose {
		logrus.Warn("Running in verbose mode")
		requestHandler = VerboseLoggingMiddleware(requestHandler)
	}

	http.ListenAndServe(serverAddress, requestHandler)
	/*
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
		log.Infof("starting on %s", serverAddress)
		err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))
	*/
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return err

}
