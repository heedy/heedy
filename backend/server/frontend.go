package server

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/server/auth"
	"github.com/spf13/afero"
)

type fContext struct {
	User   *database.User             `json:"user"`
	Routes map[string]string          `json:"routes"`
	Menu   map[string]assets.MenuItem `json:"menu"`
}

type aContext struct {
	User    *database.User    `json:"user"`
	Request *auth.AuthRequest `json:"request`
}

// FrontendMux represents the frontend
func FrontendMux(a *assets.Assets) (*chi.Mux, error) {
	mux := chi.NewMux()

	frontendFS := afero.NewBasePathFs(a.FS, "/public")

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
		ctx := r.Context().Value(auth.CTX).(*auth.Context)
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
			Routes: a.Config.Frontend.PublicRoutes,
			Menu:   a.Config.Frontend.PublicMenu,
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

	// The authorization flow (login/give permissions page)
	abytes, err := afero.ReadFile(frontendFS, "/auth.html")
	if err != nil {
		return nil, err
	}
	aTemplate, err := template.New("auth").Parse(string(abytes))
	if err != nil {
		return nil, err
	}

	mux.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(auth.CTX).(*auth.Context)
		ctx.Log.Debug("Running auth template")
		aTemplate.Execute(w, &aContext{
			User:    nil,
			Request: nil,
		})
		return
	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/static/", http.FileServer(afero.NewHttpFs(frontendFS)))

	return mux, nil
}
