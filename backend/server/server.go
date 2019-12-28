package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins"
	"github.com/heedy/heedy/backend/updater"

	"github.com/sirupsen/logrus"
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
		logrus.Warn("Running in verbose mode")
		requestHandler = VerboseLoggingMiddleware(requestHandler, nil)
	}

	err = nil
	srv := &http.Server{
		Addr:    serverAddress,
		Handler: requestHandler,
	}

	// Now load the plugins (so that the server is ready when they are loaded)
	go func() {
		logrus.Info("Initializing plugins...")
		err = pm.Start(requestHandler)
		if err != nil {
			srv.Close()
			return
		}
		logrus.Infof("Running heedy on %s", serverAddress)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		srv.Close()

		// If there is another ctrl c, kill the program
		<-c
		logrus.Error("Killing")
		go func() {
			time.Sleep(2)
			os.Exit(1)
		}()
		pm.Kill()
		os.Exit(1)
	}()

	// We add a special handler to allow restarting the server
	restartServer := false
	mux.HandleFunc("/api/server/restart", func(w http.ResponseWriter, r *http.Request) {
		db := rest.CTX(r).DB
		a := db.AdminDB().Assets()
		if db.ID() != "heedy" && !a.Config.UserIsAdmin(db.ID()) {
			rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only admins can restart heedy"))
			return
		}
		rest.CTX(r).Log.Warn("Restart requested")
		restartServer = true
		c <- os.Interrupt

		rest.WriteResult(w, r, nil)
	})

	serr := srv.ListenAndServe()
	if serr != http.ErrServerClosed {
		err = serr
	}
	logrus.Info("Stopping plugins...")
	pm.Close()
	db.Close()
	logrus.Info("Done")
	if restartServer {
		logrus.Info("Restarting")
		return updater.RunHeedy(a.FolderPath)
	}
	return err
}
