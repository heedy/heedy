package server

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/buildinfo"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/updater"

	"github.com/heedy/heedy/api/golang/rest"
)

func FireEvent(w http.ResponseWriter, r *http.Request) {
	var err error
	c := rest.CTX(r)
	if c.DB.Type() != database.AdminType {
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

func CreateApp(w http.ResponseWriter, r *http.Request) {
	var c database.App
	if err := rest.UnmarshalRequest(r, &c); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	db := rest.CTX(r).DB
	cid, _, err := db.CreateApp(&c)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	c2, err := db.ReadApp(cid, &database.ReadAppOptions{
		AccessToken: true,
	})
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
			App: c.ID,
			Event:      "app_settings_update",
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

/*
type appStruct {
	*database.App

	Unique bool `json:"unique"`
}

func GetPluginApps(w http.ResponseWriter, r *http.Request) {
	// Get all the apps available for creation
	a := rest.CTX(r).DB.AdminDB().Assets()

	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only logged in users can list plugin apps"))
		return
	}



	m := make(map[string]appStruct)


}
*/

func GetSourceScopes(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only logged in users can list scopes"))
		return
	}
	a := rest.CTX(r).DB.AdminDB().Assets()
	stype := chi.URLParam(r, "sourcetype")
	scopes, err := a.Config.GetSourceScopes(stype)
	rest.WriteJSON(w, r, scopes, err)
}

func GetAppScopes(w http.ResponseWriter, r *http.Request) {
	a := rest.CTX(r).DB.AdminDB().Assets()

	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only logged in users can list scopes"))
		return
	}
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
		"self.sources":   "Allows the app to create and manage its own sources of all types",
	}

	// Generate the source type scopes
	for stype := range a.Config.SourceTypes {
		smap[fmt.Sprintf("sources.%s", stype)] = fmt.Sprintf("All permissions for sources of type '%s'", stype)
		smap[fmt.Sprintf("sources.%s:read", stype)] = fmt.Sprintf("Read access for your sources of type '%s'", stype)
		smap[fmt.Sprintf("sources.%s:delete", stype)] = fmt.Sprintf("Can delete your sources of type '%s'", stype)

		smap[fmt.Sprintf("shared.%s", stype)] = fmt.Sprintf("All permissions for sources of type '%s' that were shared with you", stype)
		smap[fmt.Sprintf("shared.%s:read", stype)] = fmt.Sprintf("Read access for your sources of type '%s' that were shared with you", stype)

		smap[fmt.Sprintf("self.sources.%s", stype)] = fmt.Sprintf("Allows the app to create and manage its own sources of type '%s'", stype)

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
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
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
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
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
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only admins can add remove admin status"))
		return
	}
	username := chi.URLParam(r, "username")
	rest.WriteResult(w, r, a.RemAdmin(username))
}

func GetUpdates(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	ui := updater.GetInfo(a.FolderPath)
	rest.WriteJSON(w, r, ui, nil)
}

func GetConfigFile(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	b, err := updater.ReadConfigFile(a.FolderPath)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func PostConfigFile(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	defer r.Body.Close()

	//Limit requests to the limit given in configuration
	b, err := ioutil.ReadAll(io.LimitReader(r.Body, *a.Config.RequestBodyByteLimit))
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
	}
	rest.WriteResult(w, r, updater.SetConfigFile(a.FolderPath, b))
}

func PatchUConfig(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	defer r.Body.Close()

	c := assets.NewConfiguration()
	err := rest.UnmarshalRequest(r, c)
	if err == nil {
		err = updater.ModifyConfigFile(a.FolderPath, c)
	}

	rest.WriteResult(w, r, err)
}

func GetUpdateStatus(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	rest.WriteResult(w, r, updater.Status(a.FolderPath))
}

func GetUConfig(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	c, err := updater.ReadConfig(a)
	rest.WriteJSON(w, r, c, err)
}

func GetAllPlugins(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	p, err := updater.ListPlugins(a.FolderPath)
	rest.WriteJSON(w, r, p, err)
}

func PostPlugin(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	a := db.AdminDB().Assets()
	if db.Type() != database.AdminType && !a.Config.UserIsAdmin(db.ID()) {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Server settings are admin-only"))
		return
	}
	r.ParseMultipartForm(50 << 20)
	file, _, err := r.FormFile("zipfile")
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	// Upload the zip file
	tmpFile, err := ioutil.TempFile(os.TempDir(), "heedy-plugin-*.zip")
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	zipFile := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(zipFile)
	}()

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	if err = tmpFile.Close(); err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	rest.WriteResult(w, r, updater.UpdatePlugin(a.FolderPath, zipFile))
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

	v1mux.Post("/apps", CreateApp)
	v1mux.Get("/apps", ListApps)
	v1mux.Get("/apps/{appid}", ReadApp)
	v1mux.Patch("/apps/{appid}", UpdateApp)
	v1mux.Delete("/apps/{appid}", DeleteApp)

	v1mux.Get("/server/scopes/{sourcetype}", GetSourceScopes)
	v1mux.Get("/server/scopes", GetAppScopes)
	v1mux.Get("/server/version", GetVersion)

	v1mux.Get("/server/admin", GetAdminUsers)
	v1mux.Post("/server/admin/{username}", AddAdminUser)
	v1mux.Delete("/server/admin/{username}", RemoveAdminUser)

	v1mux.Get("/server/updates", GetUpdates)
	v1mux.Get("/server/updates/status", GetUpdateStatus)
	v1mux.Get("/server/updates/heedy.conf", GetConfigFile)
	v1mux.Post("/server/updates/heedy.conf", PostConfigFile)
	v1mux.Get("/server/updates/config", GetUConfig)
	v1mux.Patch("/server/updates/config", PatchUConfig)
	v1mux.Get("/server/updates/plugins", GetAllPlugins)
	v1mux.Post("/server/updates/plugins", PostPlugin)

	apiMux := chi.NewMux()
	apiMux.NotFound(APINotFound)
	apiMux.Mount("/heedy/v1", v1mux)
	return apiMux, nil
}
