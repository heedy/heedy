package server

import (
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
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
	mux.Post("/setup", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Got create request")
		sm := &setupMessage{}
		err := UnmarshalRequest(r, sm)
		if err != nil {
			// ugh
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mydir := directory
		if sm.Directory != nil {
			mydir = *sm.Directory
		}

		log.Infof("Creating database in '%s'", mydir)
		a, err := assets.Create(mydir, sm.Config, configFile)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = database.Create(a); err != nil {
			log.Error(err)
			os.RemoveAll(mydir)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		db, err := database.Open(a)
		if err != nil {
			log.Error(err)
			os.RemoveAll(mydir)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer db.Close()

		// Now add the default user
		if err = db.CreateUser(&database.User{
			Details: database.Details{
				Name: &sm.User.Name,
			},
			Password: &sm.User.Password,
		}); err != nil {
			log.Error(err)
			os.RemoveAll(mydir)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Info("Success")

		w.WriteHeader(http.StatusOK)
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

	return http.ListenAndServe(setupBind, mux)
}
