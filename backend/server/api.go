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

func ListSources(w http.ResponseWriter,r *http.Request) {
	var o database.ListSourcesOptions
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	sl,err := CTX(r).DB.ListSources(&o)
	WriteJSON(w, r, sl, err)
}

func CreateSource(w http.ResponseWriter, r *http.Request) {
	var s database.Source
	err := UnmarshalRequest(r, &s)
	if err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	adb := CTX(r).DB

	sid, err := adb.CreateSource(&s)
	if err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	s2, err := adb.ReadSource(sid, nil)

	WriteJSON(w, r, s2, err)
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

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	var c database.Connection
	if err := UnmarshalRequest(r, &c); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	db := CTX(r).DB
	cid,_, err := db.CreateConnection(&c)
	if err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	c2, err := db.ReadConnection(cid,&database.ReadConnectionOptions{
		APIKey: true,
	})
	WriteJSON(w,r,c2,err)
}

func ReadConnection(w http.ResponseWriter, r *http.Request) {
	var o database.ReadConnectionOptions
	cid := chi.URLParam(r, "connectionid")
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := CTX(r).DB.ReadConnection(cid, &o)
	WriteJSON(w, r, s, err)
}


func UpdateConnection(w http.ResponseWriter, r *http.Request) {
	var c database.Connection

	if err := UnmarshalRequest(r, &c); err != nil {
		WriteJSONError(w, r, 400, err)
		return
	}
	c.ID = chi.URLParam(r, "connectionid")
	WriteResult(w, r, CTX(r).DB.UpdateConnection(&c))
}

func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	cid := chi.URLParam(r, "connectionid")
	WriteResult(w, r, CTX(r).DB.DelConnection(cid))
}


func ListConnections(w http.ResponseWriter,r *http.Request) {
	var o database.ListConnectionOptions
	err := queryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	cl,err := CTX(r).DB.ListConnections(&o)
	WriteJSON(w, r, cl, err)
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
	v1mux.Get("/source",ListSources)
	v1mux.Get("/source/{sourceid}", ReadSource)
	v1mux.Patch("/source/{sourceid}", UpdateSource)
	v1mux.Delete("/source/{sourceid}", DeleteSource)

	v1mux.Post("/connection", CreateConnection)
	v1mux.Get("/connection", ListConnections)
	v1mux.Get("/connection/{connectionid}",ReadConnection)
	v1mux.Patch("/connection/{connectionid}",UpdateConnection)
	v1mux.Delete("/connection/{connectionid}",DeleteConnection)

	apiMux := chi.NewMux()
	apiMux.NotFound(APINotFound)
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
