package server

import (
	"html/template"
	"net/http"

	"github.com/connectordb/connectordb/src/assets"
	"github.com/connectordb/connectordb/src/database"
	"github.com/connectordb/connectordb/src/server/auth"
	"github.com/go-chi/chi"
	"github.com/spf13/afero"
)

type fContext struct {
	User   *database.User             `json:"user"`
	Routes map[string]string          `json:"routes"`
	Menu   map[string]assets.MenuItem `json:"menu"`
}

// FrontendMux represents the frontend
func FrontendMux(a *assets.Assets) (*chi.Mux, error) {
	mux := chi.NewMux()

	frontendFS := afero.NewBasePathFs(a.FS, "/public")

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
		l := ctx.Log

		u, err := ctx.DB.ThisUser()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if u == nil {
			// Not a user, so show public database
			l.Info("Running template for public")
			fTemplate.Execute(w, &fContext{
				User:   u,
				Routes: a.Config.Frontend.PublicRoutes,
				Menu:   a.Config.Frontend.PublicMenu,
			})
			return
		}
		l.Infof("Running template for %s", *u.Name)
		fTemplate.Execute(w, &fContext{
			User:   u,
			Routes: a.Config.Frontend.Routes,
			Menu:   a.Config.Frontend.Menu,
		})

	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/", http.FileServer(afero.NewHttpFs(frontendFS)))

	return mux, nil
}
