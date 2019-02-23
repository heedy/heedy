package server

import (
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/connectordb/connectordb/assets"
	"github.com/connectordb/connectordb/database"
	"github.com/go-chi/chi"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
)

type setupContext struct {
	Config    *assets.Configuration
	Directory string
}

// Setup runs the setup server. All of the arguments are optional - include empty strings
// for the directory and configFile if they are not given, and nil for configuration if no settings
// were given.
// This will load the default config, overwritten with configFile, overwritten with c, and use that
// as the "defaults" for fields given to the user.
func Setup(directory string, c *assets.Configuration, configFile string, setupBind string) error {

	setupbytes, err := afero.ReadFile(assets.FS(), "/setup/index.html")
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
	mux.Mount("/setup/", http.FileServer(afero.NewHttpFs(assets.FS())))

	// /setup is POSTed with info, and this function prepares the database
	mux.Post("/setup", func(w http.ResponseWriter, r *http.Request) {
		c := assets.NewConfiguration()
		err := UnmarshalRequest(r, c)
		if err != nil {
			// ugh
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		a, err := assets.Create(directory, c, configFile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = database.Create(a); err != nil {
			os.RemoveAll(directory)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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

	log.Infof("Open 'http://%s:%s' in your browser to set up ConnectorDB", host, port)

	return http.ListenAndServe(setupBind, mux)
}
