package server

import (
	"net/http"

	"github.com/connectordb/connectordb/src/assets"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

// The this route operates with reference to the logged in connection.
func GetThis(w http.ResponseWriter, r *http.Request) {
	//cdb := r.Context().Value("cdb").(database.DB)
	r.Context().Value(cK("log")).(*logrus.Entry).Info("Here")
	w.WriteHeader(http.StatusAccepted)
}

// APIMux gives the REST API
func APIMux(a *assets.Assets) (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/this", GetThis)

	apiMux := chi.NewMux()
	apiMux.Mount("/cdb/v1", v1mux)
	return apiMux, nil
}
