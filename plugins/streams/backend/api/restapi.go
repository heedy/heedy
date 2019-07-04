package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/server"
)



func ReadData(w http.ResponseWriter, r *http.Request) {
	c := server.CTX(r)
	c.Log.Debug("Got request!")
	w.WriteHeader(200)
	w.Write([]byte("Hi!"))
}

func DataMux() *chi.Mux {
	m := chi.NewMux()

	m.Get("/", ReadData)

	return m
}

// Handler is the global router for the stream API
var Handler = func() *chi.Mux {
	m := chi.NewMux()

	m.Mount("/data", DataMux())

	return m
}()
