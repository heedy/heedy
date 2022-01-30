package server

import (
	"errors"
	"net/http"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins"

	"github.com/heedy/heedy/api/golang/rest"
)

func ReadUser(w http.ResponseWriter, r *http.Request) {
	var o database.ReadUserOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	username, err := rest.URLParam(r, "username", err)
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
	err := rest.UnmarshalRequest(r, &u)
	u.ID, err = rest.URLParam(r, "username", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	rest.WriteResult(w, r, rest.CTX(r).DB.UpdateUser(&u))
}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	username, err := rest.URLParam(r, "username", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
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

func ReadUserSettings(w http.ResponseWriter, r *http.Request) {
	username, err := rest.URLParam(r, "username", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	v, err := rest.CTX(r).DB.ReadUserSettings(username)
	rest.WriteJSON(w, r, v, err)
}

func ReadUserPluginSettings(w http.ResponseWriter, r *http.Request) {
	username, err := rest.URLParam(r, "username", nil)
	plugin, err := rest.URLParam(r, "plugin", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	v, err := rest.CTX(r).DB.ReadUserPluginSettings(username, plugin)
	rest.WriteJSON(w, r, v, err)
}

func UpdateUserPluginSettings(w http.ResponseWriter, r *http.Request) {
	username, err := rest.URLParam(r, "username", nil)
	plugin, err := rest.URLParam(r, "plugin", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	var v map[string]interface{}

	if err := rest.UnmarshalRequest(r, &v); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}

	rest.WriteResult(w, r, rest.CTX(r).DB.UpdateUserPluginSettings(username, plugin, v))
}

func GetUserSettingSchemas(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("access_denied: Only logged in users can read preference schemas"))
		return
	}

	schemaMap := make(map[string]map[string]interface{})

	cfg := db.AdminDB().Assets().Config
	if len(cfg.UserSettingsSchema) > 0 {
		schemaMap["heedy"] = cfg.GetUserSettingsSchema()
	}
	for p, pv := range cfg.Plugins {
		if len(pv.UserSettingsSchema) > 0 {
			schemaMap[p] = pv.GetUserSettingsSchema()
		}
	}

	rest.WriteJSON(w, r, schemaMap, nil)
}

func ListUserSessions(w http.ResponseWriter, r *http.Request) {
	username, err := rest.URLParam(r, "username", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	v, err := rest.CTX(r).DB.ListUserSessions(username)
	rest.WriteJSON(w, r, v, err)
}

func DeleteUserSession(w http.ResponseWriter, r *http.Request) {
	username, err := rest.URLParam(r, "username", nil)
	sessionid, err := rest.URLParam(r, "sessionid", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	rest.WriteResult(w, r, rest.CTX(r).DB.DelUserSession(username, sessionid))
}

func ListObjects(w http.ResponseWriter, r *http.Request) {
	var o database.ListObjectsOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	sl, err := rest.CTX(r).DB.ListObjects(&o)
	rest.WriteJSON(w, r, sl, err)
}

func CreateObject(w http.ResponseWriter, r *http.Request) {
	var s database.Object
	var o database.ReadObjectOptions
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

	sid, err := adb.CreateObject(&s)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	s2, err := adb.ReadObject(sid, &o)

	rest.WriteJSON(w, r, s2, err)
}

func ReadObject(w http.ResponseWriter, r *http.Request) {
	var o database.ReadObjectOptions
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	srcid, err := rest.URLParam(r, "objectid", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := rest.CTX(r).DB.ReadObject(srcid, &o)
	rest.WriteJSON(w, r, s, err)
}

func UpdateObject(w http.ResponseWriter, r *http.Request) {
	var s database.Object
	err := rest.UnmarshalRequest(r, &s)
	s.ID, err = rest.URLParam(r, "objectid", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	rest.WriteResult(w, r, rest.CTX(r).DB.UpdateObject(&s))
}

func DeleteObject(w http.ResponseWriter, r *http.Request) {
	sid, err := rest.URLParam(r, "objectid", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	rest.WriteResult(w, r, rest.CTX(r).DB.DelObject(sid))
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
		// a plugin app, which will be auto-populated with timeseries and managed by the plugin.
		if c.Name != nil && *c.Name != "" {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("Creating a plugin app requires all app fields other than 'plugin' to be empty"))
		}
		owner := ""
		if c.Owner != nil {
			// CreateApp internally makes sure that the owner is non-empty, and valid
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
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	cid, err := rest.URLParam(r, "appid", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	s, err := rest.CTX(r).DB.ReadApp(cid, &o)
	rest.WriteJSON(w, r, s, err)
}

func UpdateApp(w http.ResponseWriter, r *http.Request) {
	var c database.App
	err := rest.UnmarshalRequest(r, &c)
	c.ID, err = rest.URLParam(r, "appid", err)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	err = rest.CTX(r).DB.UpdateApp(&c)

	// If the access token has changed, write the new one in the result
	res := struct {
		Result      string  `json:"result"`
		AccessToken *string `json:"access_token,omitempty"`
	}{Result: "ok", AccessToken: c.AccessToken}

	rest.WriteJSON(w, r, res, err)

}

func DeleteApp(w http.ResponseWriter, r *http.Request) {
	cid, err := rest.URLParam(r, "appid", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
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
