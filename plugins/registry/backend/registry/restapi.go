package registry

import (
	"github.com/go-chi/chi"
)

var Handler = func() *chi.Mux {
	m := chi.NewMux()
	return m
}()
