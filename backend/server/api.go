package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/buildinfo"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"

	"github.com/heedy/heedy/api/golang/rest"
)

func FireEvent(w http.ResponseWriter, r *http.Request) {
	var err error
	c := rest.CTX(r)
	if c.DB.ID() != "heedy" {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("access_denied: Only plugins may fire events"))
		return
	}
	var e events.Event
	if err = rest.UnmarshalRequest(r, &e); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	if err = events.FillEvent(c.DB.AdminDB(), &e); err == nil {
		events.Fire(&e)

	}
	rest.WriteResult(w, r, err)
}

func ReadUser(w http.ResponseWriter, r *http.Request) {
	var o database.ReadUserOptions
	username := chi.URLParam(r, "username")
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	u, err := rest.CTX(r).DB.ReadUser(username, &o)
	rest.WriteJSON(w, r, u, err)
}
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var u database.User

	if err := rest.UnmarshalRequest(r, &u); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	rest.WriteResult(w, r, rest.CTX(r).DB.CreateUser(&u))
}
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var u database.User

	if err := rest.UnmarshalRequest(r, &u); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	u.ID = chi.URLParam(r, "username")
	rest.WriteResult(w, r, rest.CTX(r).DB.UpdateUser(&u))
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	rest.WriteResult(w, r, rest.CTX(r).DB.DelUser(username))
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	var o database.ListUsersOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	sl, err := rest.CTX(r).DB.ListUsers(&o)
	rest.WriteJSON(w, r, sl, err)
}

func ListSources(w http.ResponseWriter, r *http.Request) {
	var o database.ListSourcesOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	sl, err := rest.CTX(r).DB.ListSources(&o)
	rest.WriteJSON(w, r, sl, err)
}

func CreateSource(w http.ResponseWriter, r *http.Request) {
	var s database.Source
	err := rest.UnmarshalRequest(r, &s)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	adb := rest.CTX(r).DB

	sid, err := adb.CreateSource(&s)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	s2, err := adb.ReadSource(sid, nil)

	rest.WriteJSON(w, r, s2, err)
}

func ReadSource(w http.ResponseWriter, r *http.Request) {
	var o database.ReadSourceOptions
	srcid := chi.URLParam(r, "sourceid")
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := rest.CTX(r).DB.ReadSource(srcid, &o)
	rest.WriteJSON(w, r, s, err)
}

func UpdateSource(w http.ResponseWriter, r *http.Request) {
	var s database.Source

	if err := rest.UnmarshalRequest(r, &s); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	s.ID = chi.URLParam(r, "sourceid")
	rest.WriteResult(w, r, rest.CTX(r).DB.UpdateSource(&s))
}

func DeleteSource(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sourceid")
	rest.WriteResult(w, r, rest.CTX(r).DB.DelSource(sid))
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	var c database.Connection
	if err := rest.UnmarshalRequest(r, &c); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	db := rest.CTX(r).DB
	cid, _, err := db.CreateConnection(&c)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	c2, err := db.ReadConnection(cid, &database.ReadConnectionOptions{
		AccessToken: true,
	})
	rest.WriteJSON(w, r, c2, err)
}

func ReadConnection(w http.ResponseWriter, r *http.Request) {
	var o database.ReadConnectionOptions
	cid := chi.URLParam(r, "connectionid")
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := rest.CTX(r).DB.ReadConnection(cid, &o)
	rest.WriteJSON(w, r, s, err)
}

func UpdateConnection(w http.ResponseWriter, r *http.Request) {
	var c database.Connection

	if err := rest.UnmarshalRequest(r, &c); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	c.ID = chi.URLParam(r, "connectionid")
	err := rest.CTX(r).DB.UpdateConnection(&c)
	if err == nil && c.Settings != nil {
		rest.CTX(r).Events.Fire(&events.Event{
			Connection: c.ID,
			Event:      "connection_settings_update",
		})
	}
	rest.WriteResult(w, r, err)

}

func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	cid := chi.URLParam(r, "connectionid")
	rest.WriteResult(w, r, rest.CTX(r).DB.DelConnection(cid))
}

func ListConnections(w http.ResponseWriter, r *http.Request) {
	var o database.ListConnectionOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	cl, err := rest.CTX(r).DB.ListConnections(&o)
	rest.WriteJSON(w, r, cl, err)
}

func GetSourceScopes(w http.ResponseWriter, r *http.Request) {
	// TODO: figure out whether to require auth for this
	a := rest.CTX(r).DB.AdminDB().Assets()
	stype := chi.URLParam(r, "sourcetype")
	scopes, err := a.Config.GetSourceScopes(stype)
	rest.WriteJSON(w, r, scopes, err)
}

func GetConnectionScopes(w http.ResponseWriter, r *http.Request) {
	a := rest.CTX(r).DB.AdminDB().Assets()
	// Now our job is to generate all of the scopes
	// TODO: language support
	// TODO: maybe require auth for this?

	var smap = map[string]string{
		"owner":          "All available access to your user",
		"owner:read":     "Read your user info",
		"owner:update":   "Modify your user's info",
		"users":          "All permissions for all users",
		"users:read":     "Read all users that you can read",
		"users:update":   "Modify info for all users you can modify",
		"sources":        "All permissions for all sources of all types",
		"sources:read":   "Read all sources belonging to you (of all types)",
		"sources:update": "Modify data of all sources belonging to you (of all types)",
		"sources:delete": "Delete any sources belonging to you (of all types)",
		"shared":         "All permissions for sources shared with you (of all types)",
		"shared:read":    "Read sources of all types that were shared with you",
		"self.sources":   "Allows the connection to create and manage its own sources of all types",
	}

	// Generate the source type scopes
	for stype := range a.Config.SourceTypes {
		smap[fmt.Sprintf("sources.%s", stype)] = fmt.Sprintf("All permissions for sources of type '%s'", stype)
		smap[fmt.Sprintf("sources.%s:read", stype)] = fmt.Sprintf("Read access for your sources of type '%s'", stype)
		smap[fmt.Sprintf("sources.%s:delete", stype)] = fmt.Sprintf("Can delete your sources of type '%s'", stype)

		smap[fmt.Sprintf("shared.%s", stype)] = fmt.Sprintf("All permissions for sources of type '%s' that were shared with you", stype)
		smap[fmt.Sprintf("shared.%s:read", stype)] = fmt.Sprintf("Read access for your sources of type '%s' that were shared with you", stype)

		smap[fmt.Sprintf("self.sources.%s", stype)] = fmt.Sprintf("Allows the connection to create and manage its own sources of type '%s'", stype)

		// And now generate the per-type scopes
		stypemap := a.Config.SourceTypes[stype].Scopes
		if stypemap != nil {
			for sscope := range *stypemap {
				smap[fmt.Sprintf("sources.%s:%s", stype, sscope)] = (*stypemap)[sscope]
				//smap[fmt.Sprintf("self.sources.%s:%s",stype,sscope)] = (*stypemap)[sscope]
				smap[fmt.Sprintf("shared.%s:%s", stype, sscope)] = (*stypemap)[sscope]
			}
		}
	}

	rest.WriteJSON(w, r, smap, nil)

}

func GetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(buildinfo.Version))
}

func GetAdminUsers(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.ID() != "heedy" && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only admins can list admins"))
		return
	}
	if a.Config.AdminUsers == nil {
		rest.WriteJSON(w, r, []string{}, nil)
		return
	}
	rest.WriteJSON(w, r, *a.Config.AdminUsers, nil)
}

func AddAdminUser(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.ID() != "heedy" && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only admins can add admin users"))
		return
	}
	username := chi.URLParam(r, "username")
	_, err := rest.CTX(r).DB.ReadUser(username, nil)
	if err == nil {
		err = a.AddAdmin(username)
	}
	rest.WriteResult(w, r, err)
}
func RemoveAdminUser(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.ID() != "heedy" && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only admins can add remove admin status"))
		return
	}
	username := chi.URLParam(r, "username")
	rest.WriteResult(w, r, a.RemAdmin(username))
}

func APINotFound(w http.ResponseWriter, r *http.Request) {
	rest.WriteJSONError(w, r, http.StatusNotFound, rest.ErrNotFound)
}

// APIMux gives the REST API
func APIMux() (*chi.Mux, error) {

	v1mux := chi.NewMux()

	v1mux.Get("/events", EventWebsocket)
	v1mux.Post("/events", FireEvent)

	v1mux.Post("/users", CreateUser)
	v1mux.Get("/users", ListUsers)
	v1mux.Get("/users/{username}", ReadUser)
	v1mux.Patch("/users/{username}", UpdateUser)
	v1mux.Delete("/users/{username}", DeleteUser)

	v1mux.Post("/sources", CreateSource)
	v1mux.Get("/sources", ListSources)
	v1mux.Get("/sources/{sourceid}", ReadSource)
	v1mux.Patch("/sources/{sourceid}", UpdateSource)
	v1mux.Delete("/sources/{sourceid}", DeleteSource)

	v1mux.Post("/connections", CreateConnection)
	v1mux.Get("/connections", ListConnections)
	v1mux.Get("/connections/{connectionid}", ReadConnection)
	v1mux.Patch("/connections/{connectionid}", UpdateConnection)
	v1mux.Delete("/connections/{connectionid}", DeleteConnection)

	v1mux.Get("/meta/scopes/{sourcetype}", GetSourceScopes)
	v1mux.Get("/meta/scopes", GetConnectionScopes)
	v1mux.Get("/meta/version", GetVersion)

	v1mux.Get("/settings/admin", GetAdminUsers)
	v1mux.Post("/settings/admin/{username}", AddAdminUser)
	v1mux.Delete("/settings/admin/{username}", RemoveAdminUser)

	apiMux := chi.NewMux()
	apiMux.NotFound(APINotFound)
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
