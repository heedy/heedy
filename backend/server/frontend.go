package server

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/spf13/afero"
)

type frontendPlugin struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type fContext struct {
	User     *database.User   `json:"user"`
	Admin    bool             `json:"admin"`
	Frontend []frontendPlugin `json:"frontend"`
}

type aContext struct {
	User    *database.User `json:"user"`
	Request *CodeRequest   `json:"request"`
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
		w.Header().Add("X-Frame-Options", "DENY")
		w.Header().Add("Cache-Control", "private, no-cache")

		ctx := rest.CTX(r)
		var u *database.User
		var err error

		if _, ok := ctx.DB.(*database.UserDB); ok {
			u, err = ctx.DB.ReadUser(ctx.DB.ID(), &database.ReadUserOptions{
				Avatar: true,
			})
			if err != nil {
				rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
				return
			}
		}

		cfg := assets.Config()

		frontendPlugins := make([]frontendPlugin, 0)
		if cfg.Frontend != nil {
			frontendPlugins = append(frontendPlugins, frontendPlugin{
				Name: "heedy",
				Path: *cfg.Frontend,
			})
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
			}
		}

		/*
			sourceMap := make(map[string]*assets.SourceTypeFrontend)
			for k, v := range cfg.SourceTypes {
				sourceMap[k] = v.Frontend
			}
		*/

		if u == nil {
			// Running template as public
			err = fTemplate.Execute(w, &fContext{
				User:     nil,
				Admin:    false,
				Frontend: frontendPlugins,
			})
			if err != nil {
				rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		err = fTemplate.Execute(w, &fContext{
			User:     u,
			Admin:    ctx.DB.AdminDB().Assets().Config.UserIsAdmin(*u.UserName),
			Frontend: frontendPlugins,
		})
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		return

	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/static/", http.FileServer(afero.NewHttpFs(frontendFS)))

	return mux, nil
}
