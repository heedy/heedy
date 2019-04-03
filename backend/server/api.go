package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/database"
)

func ReadUser(w http.ResponseWriter, r *http.Request) {
	var o database.ReadUserOptions
	username := chi.URLParam(r, "username")
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	u, err := CTX(r).DB.ReadUser(username, &o)
	WriteJSON(w, r, u, err)
}
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var u database.User

	if err := UnmarshalRequest(r, &u); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	u.ID = chi.URLParam(r, "username")
	WriteResult(w, r, CTX(r).DB.UpdateUser(&u))
}

func APINotFound(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, r, http.StatusNotFound, errors.New("not_found: The given endpoint is not available"))
}

// APIMux gives the REST API
func APIMux() (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/user/{username}", ReadUser)
	v1mux.Post("/user/{username}", UpdateUser)

	apiMux := chi.NewMux()
	apiMux.NotFound(APINotFound)
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
