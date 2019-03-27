package server

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/spf13/afero"
)

type fContext struct {
	User   *database.User             `json:"user"`
	Routes map[string]string          `json:"routes"`
	Menu   map[string]assets.MenuItem `json:"menu"`
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

		ctx := CTX(r)
		/*
			u, err := ctx.DB.ThisUser()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if u == nil {
		*/
		// Not a user, so show public database
		ctx.Log.Debug("Running template for public")
		fTemplate.Execute(w, &fContext{
			User:   nil,
			Routes: assets.Config().Frontend.PublicRoutes,
			Menu:   assets.Config().Frontend.PublicMenu,
		})
		return
		/*
			}
			l.Infof("Running template for %s", *u.Name)
			fTemplate.Execute(w, &fContext{
				User:   u,
				Routes: a.Config.Frontend.Routes,
				Menu:   a.Config.Frontend.Menu,
			})
		*/

	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/static/", http.FileServer(afero.NewHttpFs(frontendFS)))

	return mux, nil
}
