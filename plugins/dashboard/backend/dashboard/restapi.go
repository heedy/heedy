package dashboard

import "github.com/go-chi/chi"

// Handler is the global router for the timeseries API
var Handler = func() *chi.Mux {
	m := chi.NewMux()

	return m
}()
