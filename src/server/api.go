package server

import (
	"net/http"

	"github.com/connectordb/connectordb/src/assets"
	"github.com/go-chi/chi"
)

// The this route operates with reference to the logged in connection.
func GetThis(w http.ResponseWriter, r *http.Request) {
	//cdb := r.Context().Value("cdb").(database.DB)
	//r.Context().Value(cK("log")).(*logrus.Entry).Info("Here")
	w.WriteHeader(http.StatusAccepted)
}

// RequestToken permits a user to log in. It also leaves us open to future
// oauth implementation
func RequestToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {

	}
}

// APIMux gives the REST API
func APIMux(a *assets.Assets) (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/this", GetThis)
	v1mux.Post("/token", RequestToken)

	apiMux := chi.NewMux()
	apiMux.Mount("/cdb/v1", v1mux)
	return apiMux, nil
}
