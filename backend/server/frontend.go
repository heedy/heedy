package server

import (
	"errors"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/lpar/gzipped/v2"
	"github.com/spf13/afero"
)

type frontendPlugin struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type fContext struct {
	User        *database.User                    `json:"user"`
	Preferences map[string]map[string]interface{} `json:"preferences"`
	Admin       bool                              `json:"admin"`
	Plugins     []frontendPlugin                  `json:"plugins"`
	Preload     []string                          `json:"preload"`
	Verbose     bool                              `json:"verbose"`
}

type aContext struct {
	User    *database.User `json:"user"`
	Request *CodeRequest   `json:"request"`
}

// withExists allows to use lpar/gzipped
type withExists struct {
	*afero.HttpFs
}

func (we withExists) Exists(name string) bool {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return false
	}

	fullName := filepath.FromSlash(path.Clean("/" + name))
	_, err := we.Stat(fullName)
	return err == nil
}

// FrontendMux represents the frontend
func FrontendMux() (*chi.Mux, error) {
	mux := chi.NewMux()

	frontendFS := afero.NewBasePathFs(assets.Get().FS, "/public")

	// The main frontend app

	fbytes, err := afero.ReadFile(frontendFS, "/index.html")
	if err != nil {
		return nil, err
	}
	fTemplate, err := template.New("frontend").Parse(string(fbytes))
	if err != nil {
		return nil, err
	}

	// This is the main function that sets up the frontend template
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Disallow clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "private, no-cache")

		ctx := rest.CTX(r)
		var u *database.User
		var err error

		if _, ok := ctx.DB.(*database.UserDB); ok {
			u, err = ctx.DB.ReadUser(ctx.DB.ID(), &database.ReadUserOptions{
				Icon: true,
			})
			if err != nil {
				rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
				return
			}
		}

		cfg := assets.Config()

		frontendPlugins := make([]frontendPlugin, 0)
		preloads := make([]string, 0)
		if cfg.Frontend != nil {
			frontendPlugins = append(frontendPlugins, frontendPlugin{
				Name: "heedy",
				Path: *cfg.Frontend,
			})
			preloads = append(preloads, *cfg.Frontend)

		}
		if cfg.Preload != nil {
			preloads = append(preloads, (*cfg.Preload)...)
		}
		for _, p := range cfg.GetActivePlugins() {
			v, ok := cfg.Plugins[p]
			if !ok {
				rest.WriteJSONError(w, r, http.StatusInternalServerError, errors.New("Failed to find plugin in configuration"))
				return
			}
			if v.Frontend != nil {
				frontendPlugins = append(frontendPlugins, frontendPlugin{
					Name: p,
					Path: *v.Frontend,
				})
				preloads = append(preloads, *v.Frontend)
			}
			if v.Preload != nil {
				preloads = append(preloads, (*v.Preload)...)
			}
		}

		/*
			objectMap := make(map[string]*assets.ObjectTypeFrontend)
			for k, v := range cfg.ObjectTypes {
				objectMap[k] = v.Frontend
			}
		*/

		if u == nil {
			// Running template as public
			err = fTemplate.Execute(w, &fContext{
				User:    nil,
				Admin:   false,
				Plugins: frontendPlugins,
				Preload: preloads,
				Verbose: cfg.Verbose,
			})
			if err != nil {
				rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		pref, err := ctx.DB.ReadUserPreferences(*u.UserName)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}

		err = fTemplate.Execute(w, &fContext{
			User:        u,
			Preferences: pref,
			Admin:       ctx.DB.AdminDB().Assets().Config.UserIsAdmin(*u.UserName),
			Plugins:     frontendPlugins,
			Preload:     preloads,
			Verbose:     cfg.Verbose,
		})
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		return

	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/static/", gzipped.FileServer(withExists{afero.NewHttpFs(frontendFS)}))

	return mux, nil
}
