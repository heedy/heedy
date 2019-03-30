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
	Scopes []string                   `json:"scopes"`
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
		u, err := ctx.DB.User()
		if err != nil {
			WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}
		if u == nil {
			scopes, err := ctx.DB.AdminDB().GetGroupScopes("public")
			if err != nil {
				WriteJSONError(w, r, http.StatusInternalServerError, err)
				return
			}
			// Running template as public
			err = fTemplate.Execute(w, &fContext{
				User:   nil,
				Scopes: scopes,
				Routes: assets.Config().Frontend.PublicRoutes,
				Menu:   assets.Config().Frontend.PublicMenu,
			})
			if err != nil {
				WriteJSONError(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		scopes, err := ctx.DB.GetUserScopes(u.ID)
		if err != nil {
			WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}
		err = fTemplate.Execute(w, &fContext{
			User:   u,
			Scopes: scopes,
			Routes: assets.Config().Frontend.Routes,
			Menu:   assets.Config().Frontend.Menu,
		})
		if err != nil {
			WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		return

	})

	// Handles getting all assets other than the root webpage
	mux.Mount("/static/", http.FileServer(afero.NewHttpFs(frontendFS)))

	return mux, nil
}
