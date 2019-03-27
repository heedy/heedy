package server

import (
	"net/http"

	"github.com/go-chi/chi"
)

// The this route operates with reference to the logged in connection.
func GetThis(w http.ResponseWriter, r *http.Request) {
	//heedy := r.Context().Value("heedy").(database.DB)
	//r.Context().Value(cK("log")).(*logrus.Entry).Info("Here")
	w.WriteHeader(http.StatusAccepted)
}

// APIMux gives the REST API
func APIMux() (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/this", GetThis)

	apiMux := chi.NewMux()
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
