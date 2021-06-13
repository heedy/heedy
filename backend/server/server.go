package server

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
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
	db, err := database.Open(a)
	if err != nil {
		return err
	}

	auth := NewAuth(db)

	apiMux, err := APIMux()
	if err != nil {
		db.Close()
		return err
	}
	authMux, err := AuthMux(auth)
	if err != nil {
		db.Close()
		return err
	}
	fMux, err := FrontendMux()
	if err != nil {
		db.Close()
		return err
	}

	mux := chi.NewMux()
	mux.Mount("/api", apiMux)
	mux.Mount("/auth", authMux)
	mux.Mount("/", fMux)

	pm, err := plugins.NewPluginManager(db, http.Handler(mux))
	if err != nil {
		db.Close()
		return err
	}

	requestHandler := http.Handler(NewRequestHandler(auth, pm))

	if a.Config.Verbose {
		logrus.Warn("Running in verbose mode")
		requestHandler = VerboseLoggingMiddleware(requestHandler, nil)
	}

	err = nil

	apiAddress := a.Config.GetAPI()
	apisrv := &http.Server{
		Addr:    apiAddress,
		Handler: requestHandler,
	}
	serverAddress := a.Config.GetAddr()
	srv := &http.Server{
		Addr: serverAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(r.Header.Get("X-Heedy-Key")) > 0 {
				rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("access_denied: Plugins must use backend API socket"))
				return
			}
			requestHandler.ServeHTTP(w, r)
		}),
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		// Close the frontend server, but don't yet close the API server,
		// so that plugins can cleanly exit
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
	applyUpdates := false
	mux.HandleFunc("/api/server/restart", func(w http.ResponseWriter, r *http.Request) {
		db := rest.CTX(r).DB
		a := db.AdminDB().Assets()
		if db.ID() != "heedy" && !a.Config.UserIsAdmin(db.ID()) {
			rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only admins can restart heedy"))
			return
		}
		rest.CTX(r).Log.Warn("Restart requested")
		restartServer = true
		applyUpdates = true
		c <- os.Interrupt

		rest.WriteResult(w, r, nil)
	})

	// Now start the plugin API server in one thread, and load the plugins in another,
	// after which open the listening socket

	var apisrvl net.Listener
	if strings.HasPrefix(apiAddress, "unix:") {
		apisrvl, err = net.Listen("unix", apiAddress[5:])
	} else {
		apisrvl, err = net.Listen("tcp", apiAddress)
	}
	if err != nil {
		db.Close()
		return err
	}
	go func() {
		logrus.Debugf("Running heedy plugin API on %s", serverAddress)
		apiserr := apisrv.Serve(apisrvl)
		if apiserr != http.ErrServerClosed {
			err = apiserr
			srv.Close()
			logrus.Errorf("Plugin API Server Error: %s", apiserr)
		} else {
			logrus.Debug("Plugin API server closed")
		}

	}()

	// Let the plugin API start
	runtime.Gosched()

	logrus.Info("Initializing plugins...")
	err = pm.Start(requestHandler)
	if err != nil {
		apisrv.Close()
		db.Close()
		return err
	}

	// Only start listening once the plugins are all loaded
	logrus.Infof("Running heedy on %s", serverAddress)
	var srvl net.Listener
	if strings.HasPrefix(serverAddress, "unix:") {
		srvl, err = net.Listen("unix", serverAddress[5:])
	} else {
		srvl, err = net.Listen("tcp", serverAddress)
	}
	if err != nil {
		logrus.Errorf("Error starting heedy: %s", err)
		pm.Close()
		apisrv.Close()
		db.Close()
		return err
	}

	serr := srv.Serve(srvl)
	if serr != http.ErrServerClosed {
		err = serr
	}
	logrus.Info("Stopping plugins...")
	pm.Close()
	apisrv.Close()
	db.Close()
	logrus.Info("Done")
	if restartServer {
		logrus.Info("Restarting")
		if applyUpdates {
			return updater.StartHeedy(a.FolderPath, true, "--update")
		}
		return updater.StartHeedy(a.FolderPath, true)
	}
	return err
}
