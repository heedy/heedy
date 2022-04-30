package server

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/buildinfo"
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
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
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
	revertUpdates := false
	var restartMutex sync.Mutex
	mux.Post("/api/server/restart", func(w http.ResponseWriter, r *http.Request) {
		restartMutex.Lock()
		defer func() {
			if !restartServer {
				restartMutex.Unlock()
			}
		}()
		db := rest.CTX(r).DB
		a := db.AdminDB().Assets()
		if db.ID() != "heedy" && !a.Config.UserIsAdmin(db.ID()) {
			rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("only admins can restart heedy"))
			return
		}

		var restartOptions struct {
			Backup bool `json:"backup,omitempty"`
			Update bool `json:"update,omitempty"`
			Revert bool `json:"revert,omitempty"`
		}
		if r.Body != nil {
			err = rest.UnmarshalRequest(r, &restartOptions)
			if err != nil {
				rest.WriteJSONError(w, r, http.StatusBadRequest, err)
				return
			}
		}
		if (restartOptions.Backup || restartOptions.Update) && restartOptions.Revert {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("cannot update and revert at the same time"))
			return
		}

		if restartOptions.Backup {
			updater.EnableDataBackup(db.AdminDB().Assets().FolderPath)
		} else if restartOptions.Revert && updater.GetBackupCount(db.AdminDB().Assets().FolderPath) == 0 {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("no backup to revert"))
			return
		}

		restartServer = true
		applyUpdates = restartOptions.Update || restartOptions.Backup
		revertUpdates = restartOptions.Revert

		if applyUpdates {
			rest.CTX(r).Log.Warn("Restart requested with update")
		} else if revertUpdates {
			rest.CTX(r).Log.Warn("Restart requested with update revert")
		} else {
			rest.CTX(r).Log.Warn("Restart requested")
		}
		c <- os.Interrupt

		rest.WriteResult(w, r, nil)
	})

	// Now start the plugin API server in one thread, and load the plugins in another,
	// after which open the listening socket

	var apisrvl net.Listener
	if strings.HasPrefix(apiAddress, "unix:") {
		err = os.RemoveAll(apiAddress[5:]) // Make sure the socket is removed it
		if err == nil {
			apisrvl, err = net.Listen("unix", apiAddress[5:])
		}
	} else {
		apisrvl, err = net.Listen("tcp", apiAddress)
	}
	if err != nil {
		db.Close()
		return err
	}
	go func() {
		logrus.Debugf("Running heedy plugin API on %s", apiAddress)
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
	logrus.Infof("Running heedy v%s on %s", buildinfo.Version, serverAddress)
	var srvl net.Listener
	if strings.HasPrefix(serverAddress, "unix:") {
		err = os.RemoveAll(serverAddress[5:])
		if err == nil {
			srvl, err = net.Listen("unix", serverAddress[5:])
		}
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
		} else if revertUpdates {
			return updater.StartHeedy(a.FolderPath, true, "--revert")
		}
		return updater.StartHeedy(a.FolderPath, true)
	}
	return err
}
