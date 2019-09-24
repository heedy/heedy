package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
)

// Handler is the main API handler
var Handler = func() *chi.Mux {
	m := chi.NewMux()

	m.NotFound(func(w http.ResponseWriter, r *http.Request) {
		rest.WriteJSONError(w, r, http.StatusNotFound, rest.ErrNotFound)
	})
	return m
}()
