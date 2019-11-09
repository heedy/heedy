package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins"

	"github.com/heedy/heedy/api/golang/rest"
)

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
	var o database.ReadSourceOptions
	err := rest.UnmarshalRequest(r, &s)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	err = rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	adb := rest.CTX(r).DB

	sid, err := adb.CreateSource(&s)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	s2, err := adb.ReadSource(sid, &o)

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

func CreateApp(w http.ResponseWriter, r *http.Request) {
	var c database.App
	var o database.ReadAppOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	err = rest.UnmarshalRequest(r, &c)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	db := rest.CTX(r).DB
	var cid string
	if c.Plugin == nil || *c.Plugin == "" {
		cid, _, err = db.CreateApp(&c)
	} else {
		// There is a plugin set. This means that the user might want to create
		// a plugin app, which will be auto-populated with streams and managed by the plugin.
		if c.Name != nil && *c.Name != "" {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("Creating a plugin app requires all app fields other than 'plugin' to be empty"))
		}
		owner := ""
		if c.Owner != nil {
			// CreateApp internally makes sure that
			owner = *c.Owner

		}
		cid, _, err = plugins.CreateApp(rest.CTX(r), owner, *c.Plugin)
	}

	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	c2, err := db.ReadApp(cid, &o)
	rest.WriteJSON(w, r, c2, err)
}

func ReadApp(w http.ResponseWriter, r *http.Request) {
	var o database.ReadAppOptions
	cid := chi.URLParam(r, "appid")
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := rest.CTX(r).DB.ReadApp(cid, &o)
	rest.WriteJSON(w, r, s, err)
}

func UpdateApp(w http.ResponseWriter, r *http.Request) {
	var c database.App

	if err := rest.UnmarshalRequest(r, &c); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	c.ID = chi.URLParam(r, "appid")
	err := rest.CTX(r).DB.UpdateApp(&c)
	if err == nil && c.Settings != nil {
		rest.CTX(r).Events.Fire(&events.Event{
			App:   c.ID,
			Event: "app_settings_update",
		})
	}
	rest.WriteResult(w, r, err)

}

func DeleteApp(w http.ResponseWriter, r *http.Request) {
	cid := chi.URLParam(r, "appid")
	rest.WriteResult(w, r, rest.CTX(r).DB.DelApp(cid))
}

func ListApps(w http.ResponseWriter, r *http.Request) {
	var o database.ListAppOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	cl, err := rest.CTX(r).DB.ListApps(&o)
	rest.WriteJSON(w, r, cl, err)
}
