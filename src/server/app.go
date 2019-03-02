package server

import (
	"html/template"
	"net/http"

	"github.com/connectordb/connectordb/src/assets"
	"github.com/connectordb/connectordb/src/database"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type appContext struct {
	User   *database.User             `json:"user"`
	Routes map[string]string          `json:"routes"`
	Menu   map[string]assets.MenuItem `json:"menu"`
}

// AppMux represents the app
func AppMux(a *assets.Assets) (*chi.Mux, error) {
	mux := chi.NewMux()

	appbytes, err := afero.ReadFile(a.FS, "/app/index.html")
	if err != nil {
		return nil, err
	}
	appTemplate, err := template.New("app").Parse(string(appbytes))
	if err != nil {
		return nil, err
	}

	// This is the main function that sets up the app template
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		l := r.Context().Value(cK("log")).(*logrus.Entry)
		db := r.Context().Value(cK("cdb")).(database.DB)

		u, err := db.ThisUser()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if u == nil {
			// Not a user, so show public database
			l.Info("Running template for public")
			appTemplate.Execute(w, &appContext{
				User:   u,
				Routes: a.Config.App.PublicRoutes,
				Menu:   a.Config.App.PublicMenu,
			})
			return
		}
		l.Infof("Running template for %s", *u.Name)
		appTemplate.Execute(w, &appContext{
			User:   u,
			Routes: a.Config.App.Routes,
			Menu:   a.Config.App.Menu,
		})

	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/", http.FileServer(afero.NewHttpFs(a.FS)))

	return mux, nil
}
