package server

import (
	"context"
	"errors"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/spf13/afero"

	"github.com/sirupsen/logrus"

	_ "github.com/heedy/heedy/backend/events"
)

type SetupUser struct {
	UserName string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// This is the context that is passed to creator function
type SetupContext struct {
	assets.CreateOptions
	User SetupUser `json:"user,omitempty"`
}

func SetupCreate(sc SetupContext) error {
	if sc.User.UserName == "" || sc.User.Password == "" {
		return errors.New("A default username and password is required to create a heedy database")
	}
	// Make sure the user in context is added to admin users
	sc.Config.AdminUsers = &[]string{sc.User.UserName}

	logrus.Infof("Creating database in '%s'", sc.Directory)
	a, err := assets.Create(sc.CreateOptions)
	if err != nil {
		return err
	}

	a.Config.Verbose = sc.Config.Verbose
	assets.SetGlobal(a) // Set global assets

	if err = database.Create(a); err != nil {
		os.RemoveAll(sc.Directory)
		return err
	}

	db, err := database.Open(a)
	if err != nil {
		os.RemoveAll(sc.Directory)
		return err
	}

	// An empty-string username isn't allowed, but users with unset names are allowed when creating
	dbdetails := database.Details{}
	if strings.TrimSpace(sc.User.Name) != "" {
		dbdetails.Name = &sc.User.Name
	}

	// Now add the default user
	if err = db.CreateUser(&database.User{
		Details:  dbdetails,
		UserName: &sc.User.UserName,
		Password: &sc.User.Password,
	}); err != nil {
		db.Close()
		os.RemoveAll(sc.Directory)
		return err
	}

	return db.Close()
}

// Setup runs the setup server. All of the arguments are optional - include empty strings
// for the directory and configFile if they are not given, and nil for configuration if no settings
// were given.
// This will load the default config, overwritten with configFile, overwritten with c, and use that
// as the "defaults" for fields given to the user.
func Setup(sc SetupContext) error {

	frontendFS := afero.NewBasePathFs(assets.FS(), "/public")
	setupbytes, err := afero.ReadFile(frontendFS, "/setup.html")
	if err != nil {
		return err
	}
	setupTemplate, err := template.New("setup").Parse(string(setupbytes))
	if err != nil {
		return err
	}

	fullConfig := assets.Config()

	if sc.ConfigFile != "" {
		cF, err := assets.LoadConfigFile(sc.ConfigFile)
		if err != nil {
			return err
		}
		fullConfig = assets.MergeConfig(fullConfig, cF)
	}
	if sc.Config != nil {
		fullConfig = assets.MergeConfig(fullConfig, sc.Config)
	}

	directory, err := filepath.Abs(sc.Directory)
	if err != nil {
		return err
	}
	sc.Directory = directory

	if err = assets.EnsureEmptyDatabaseFolder(directory); err != nil {
		return err
	}

	mux := chi.NewMux()

	setupServer := &http.Server{
		Addr: *fullConfig.Addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestStart := time.Now()
			logger := rest.RequestLogger(r)
			mux.ServeHTTP(w, r)
			logger.Debugf("%v", time.Since(requestStart))
		}),
	}

	if fullConfig.Verbose {
		logrus.Warn("Running in verbose mode")
		setupServer.Handler = VerboseLoggingMiddleware(setupServer.Handler, nil)
	}

	mux.Get("/setup", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup/", http.StatusFound)
	})
	mux.Get("/setup/", func(w http.ResponseWriter, r *http.Request) {
		setupTemplate.Execute(w, &sc)
	})
	// Serve the static files, but without any caching support because serviceworker is not
	// supported during setup
	mux.Mount("/static/", NewStaticHandler(afero.NewHttpFs(frontendFS), true))

	// /setup is POSTed with info, and this function prepares the database
	setupMutex := sync.Mutex{}
	setupSuccess := false
	mux.Post("/setup", func(w http.ResponseWriter, r *http.Request) {
		setupMutex.Lock()
		defer setupMutex.Unlock()
		if setupSuccess {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("Setup is already complete"))
			return
		}
		logrus.Debug("Processing database creation request")

		var scn struct {
			SetupUser
			Addr string `json:"addr,omitempty"`
			URL  string `json:"url,omitempty"`
		}
		err := rest.UnmarshalRequest(r, &scn)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			return
		}
		sc.User = scn.SetupUser
		if (scn.Addr != "" || scn.URL != "") && sc.Config == nil {
			sc.Config = assets.NewConfiguration()
		}

		if scn.Addr != "" {
			sc.Config.Addr = &scn.Addr
		}
		if scn.URL != "" {
			sc.Config.URL = &scn.URL
		}

		err = SetupCreate(sc)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			return
		}

		// Now we load the main server, as if run was called, using the same
		// logging as during setup
		runconf := assets.NewConfiguration()
		runconf.LogDir = sc.Config.LogDir
		runconf.LogLevel = sc.Config.LogLevel
		a, err := assets.Open(sc.Directory, runconf)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			os.RemoveAll(sc.Directory)
			return
		}
		a.Config.Verbose = sc.Config.Verbose
		// Reset the global assets
		assets.SetGlobal(a)

		// OK, setup was successfully completed.
		setupSuccess = true
		logrus.Info("Database Created!")
		w.WriteHeader(http.StatusOK)

		go setupServer.Shutdown(context.TODO())

	})

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup/#", http.StatusFound)
	})
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup/#", http.StatusTemporaryRedirect)
	})

	host, port, err := net.SplitHostPort(*fullConfig.Addr)
	if err != nil {
		return err
	}

	if host == "" {
		host = "localhost"
	}

	logrus.Infof("Open 'http://%s:%s' in your browser to set up heedy", host, port)

	err = setupServer.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	setupMutex.Lock()
	defer setupMutex.Unlock()
	if !setupSuccess {
		return err
	}
	logrus.Info("Running Heedy Server")
	a := assets.Get()
	defer a.Close()
	// Setup was successful. Run the full server
	return Run(a, nil)
}
