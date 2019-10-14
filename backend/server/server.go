package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins"

	log "github.com/sirupsen/logrus"
)

// RunOptions give special options for running
type RunOptions struct {
}

func Run(a *assets.Assets, o *RunOptions) error {
	//return fmt.Errorf("TESTING REVERT")
	db, err := database.Open(a)
	if err != nil {
		return err
	}

	auth := NewAuth(db)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.GetHost(), a.Config.GetPort())

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

	pm, err := plugins.NewPluginManager(db, http.Handler(mux))
	if err != nil {
		return err
	}

	requestHandler := http.Handler(NewRequestHandler(auth, pm))

	if a.Config.Verbose {
		log.Warn("Running in verbose mode")
		requestHandler = VerboseLoggingMiddleware(requestHandler, nil)
	}

	err = nil
	srv := &http.Server{
		Addr:    serverAddress,
		Handler: requestHandler,
	}

	// Now load the plugins (so that the server is ready when they are loaded)
	go func() {
		err = pm.Reload()
		if err != nil {
			srv.Close()
			return
		}
		log.Infof("Running heedy on %s", serverAddress)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Info("Cleanup...")
			srv.Close()
			log.Info("Done")
			return
		}
	}()

	serr := srv.ListenAndServe()
	if serr != http.ErrServerClosed {
		err = serr
	}
	pm.Close()
	return err
}
