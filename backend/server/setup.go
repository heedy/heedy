package server

import (
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"errors"
	"sync"
	"context"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"

	_ "github.com/heedy/heedy/backend/events"
)

type setupContext struct {
	Config    *assets.Configuration
	Directory string
}

// Message sent to user creation stuff
type setupMessage struct {
	Config    *assets.Configuration `json:"config,omitempty"`
	Directory *string               `json:"directory,omitempty"`
	User      struct {
		UserName string `json:"username"`
		Name     string `json:"name"`
		Password string `json:"password"`
	} `json:"user,omitempty"`
}

// Setup runs the setup server. All of the arguments are optional - include empty strings
// for the directory and configFile if they are not given, and nil for configuration if no settings
// were given.
// This will load the default config, overwritten with configFile, overwritten with c, and use that
// as the "defaults" for fields given to the user.
func Setup(directory string, c *assets.Configuration, configFile string, setupBind string) error {
	

	frontendFS := afero.NewBasePathFs(assets.FS(), "/public")
	setupbytes, err := afero.ReadFile(frontendFS, "/setup.html")
	if err != nil {
		return err
	}
	setupTemplate, err := template.New("setup").Parse(string(setupbytes))

	fullConfig := assets.Config()
	if configFile != "" {
		cF, err := assets.LoadConfigFile(configFile)
		if err != nil {
			return err
		}
		fullConfig = assets.MergeConfig(fullConfig, cF)
	}
	if c != nil {
		fullConfig = assets.MergeConfig(fullConfig, c)
	}

	if directory != "" {
		directory, err = filepath.Abs(directory)
		if err != nil {
			return err
		}
	}
	mux := chi.NewMux()

	setupServer := &http.Server{
		Addr: setupBind,
		Handler: mux,
	}

	
	mux.Get("/setup", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup/", http.StatusFound)
	})
	mux.Get("/setup/", func(w http.ResponseWriter, r *http.Request) {
		setupTemplate.Execute(w, &setupContext{
			Config:    fullConfig,
			Directory: directory,
		})
	})
	mux.Mount("/static/", http.FileServer(afero.NewHttpFs(frontendFS)))

	// /setup is POSTed with info, and this function prepares the database
	setupMutex := sync.Mutex{}
	setupSuccess := false
	mux.Post("/setup", func(w http.ResponseWriter, r *http.Request) {
		setupMutex.Lock()
		defer setupMutex.Unlock()
		if setupSuccess {
			rest.WriteJSONError(w,r,http.StatusBadRequest,errors.New("Setup is already complete"))
			return
		}
		log.Debug("Got create request")
		sm := &setupMessage{}
		err := rest.UnmarshalRequest(r, sm)
		if err != nil {
			rest.WriteJSONError(w,r,http.StatusBadRequest,err)
			return
		}
		if sm.Directory != nil {
			rest.WriteJSONError(w,r,http.StatusBadRequest,errors.New("The directory cannot be set from the web setup for security purposes"))
			return
		}

		// Add the user to admins
		if sm.Config.AdminUsers == nil || len(*sm.Config.AdminUsers) == 0 {
			sm.Config.AdminUsers = &[]string{sm.User.Name}
		}

		log.Infof("Creating database in '%s'", directory)
		a, err := assets.Create(directory, sm.Config, configFile)
		if err != nil {
			rest.WriteJSONError(w,r,http.StatusBadRequest,err)
			return
		}
		if err = database.Create(a); err != nil {
			os.RemoveAll(directory)
			rest.WriteJSONError(w,r,http.StatusBadRequest,err)
			return
		}

		db, err := database.Open(a)
		if err != nil {
			rest.WriteJSONError(w,r,http.StatusInternalServerError,err)
			os.RemoveAll(directory)
			return
		}
		

		// Now add the default user
		if err = db.CreateUser(&database.User{
			Details: database.Details{
				Name: &sm.User.Name,
			},
			UserName: &sm.User.UserName,
			Password: &sm.User.Password,
		}); err != nil {
			rest.WriteJSONError(w,r,http.StatusBadRequest,err)
			db.Close()
			os.RemoveAll(directory)
			return
		}

		db.Close()

		// Now we load the main server, as if run was called
		a, err = assets.Open(directory, nil)
		if err != nil {
			rest.WriteJSONError(w,r,http.StatusBadRequest,err)
			os.RemoveAll(directory)
		}
		assets.SetGlobal(a)

		// OK, setup was successfully completed.
		setupSuccess = true
		log.Info("Database Created!")
		w.WriteHeader(http.StatusOK)

		go setupServer.Shutdown(context.TODO())
		return

	})

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup/", http.StatusFound)
	})

	
	host, port, err := net.SplitHostPort(setupBind)
	if err != nil {
		return err
	}

	if host == "" {
		host = "localhost"
	}

	log.Infof("Open 'http://%s:%s' in your browser to set up heedy", host, port)

	err = setupServer.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	setupMutex.Lock()
	defer setupMutex.Unlock()
	if !setupSuccess {
		return err
	}
	log.Info("Running Heedy Server")
	// Setup was successful. Run the full server
	return Run(nil)
}
