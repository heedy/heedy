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

func CreateSource(w http.ResponseWriter, r *http.Request) {
	var s database.Source
	err := UnmarshalRequest(r, &s)
	if err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}

	var sres struct {
		ID string `json:"id"`
	}

	sres.ID, err = CTX(r).DB.CreateSource(&s)

	WriteJSON(w, r, &s, err)
}

func ReadSource(w http.ResponseWriter, r *http.Request) {
	var o database.ReadSourceOptions
	srcid := chi.URLParam(r, "sourceid")
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := CTX(r).DB.ReadSource(srcid, &o)
	WriteJSON(w, r, s, err)
}

func UpdateSource(w http.ResponseWriter, r *http.Request) {
	var s database.Source

	if err := UnmarshalRequest(r, &s); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	s.ID = chi.URLParam(r, "sourceid")
	WriteResult(w, r, CTX(r).DB.UpdateSource(&s))
}

func DeleteSource(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sourceid")
	WriteResult(w, r, CTX(r).DB.DelSource(sid))
}

func APINotFound(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, r, http.StatusNotFound, errors.New("not_found: The given endpoint is not available"))
}

// APIMux gives the REST API
func APIMux() (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/user/{username}", ReadUser)
	v1mux.Patch("/user/{username}", UpdateUser)

	v1mux.Post("/source", CreateSource)
	v1mux.Get("/source/{sourceid}", ReadSource)
	v1mux.Patch("/source/{sourceid}", UpdateSource)
	v1mux.Delete("/source/{sourceid}", DeleteSource)

	apiMux := chi.NewMux()
	apiMux.NotFound(APINotFound)
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
