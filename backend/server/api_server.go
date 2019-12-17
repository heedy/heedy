package server

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/buildinfo"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins"
	"github.com/heedy/heedy/backend/updater"
)

func GetObjectScope(w http.ResponseWriter, r *http.Request) {
	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only logged in users can list scope"))
		return
	}
	a := rest.CTX(r).DB.AdminDB().Assets()
	stype := chi.URLParam(r, "objecttype")
	scope, err := a.Config.GetObjectScope(stype)
	rest.WriteJSON(w, r, scope, err)
}

func GetAppScope(w http.ResponseWriter, r *http.Request) {
	a := rest.CTX(r).DB.AdminDB().Assets()

	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only logged in users can list scope"))
		return
	}
	// Now our job is to generate all of the scope
	// TODO: language support
	// TODO: maybe require auth for this?

	var smap = map[string]string{
		"owner":          "All available access to your user",
		"owner:read":     "Read your user info",
		"owner:update":   "Modify your user's info",
		"users":          "All permissions for all users",
		"users:read":     "Read all users that you can read",
		"users:update":   "Modify info for all users you can modify",
		"objects":        "All permissions for all objects of all types",
		"objects:read":   "Read all objects belonging to you (of all types)",
		"objects:update": "Modify data of all objects belonging to you (of all types)",
		"objects:delete": "Delete any objects belonging to you (of all types)",
		"shared":         "All permissions for objects shared with you (of all types)",
		"shared:read":    "Read objects of all types that were shared with you",
		"self.objects":   "Allows the app to create and manage its own objects of all types",
	}

	// Generate the object type scope
	for stype := range a.Config.ObjectTypes {
		smap[fmt.Sprintf("objects.%s", stype)] = fmt.Sprintf("All permissions for objects of type '%s'", stype)
		smap[fmt.Sprintf("objects.%s:read", stype)] = fmt.Sprintf("Read access for your objects of type '%s'", stype)
		smap[fmt.Sprintf("objects.%s:delete", stype)] = fmt.Sprintf("Can delete your objects of type '%s'", stype)

		smap[fmt.Sprintf("shared.%s", stype)] = fmt.Sprintf("All permissions for objects of type '%s' that were shared with you", stype)
		smap[fmt.Sprintf("shared.%s:read", stype)] = fmt.Sprintf("Read access for your objects of type '%s' that were shared with you", stype)

		smap[fmt.Sprintf("self.objects.%s", stype)] = fmt.Sprintf("Allows the app to create and manage its own objects of type '%s'", stype)

		// And now generate the per-type scope
		stypemap := a.Config.ObjectTypes[stype].Scope
		if stypemap != nil {
			for sscope := range *stypemap {
				smap[fmt.Sprintf("objects.%s:%s", stype, sscope)] = (*stypemap)[sscope]
				//smap[fmt.Sprintf("self.objects.%s:%s",stype,sscope)] = (*stypemap)[sscope]
				smap[fmt.Sprintf("shared.%s:%s", stype, sscope)] = (*stypemap)[sscope]
			}
		}
	}

	rest.WriteJSON(w, r, smap, nil)

}

type pluginApp struct {
	*database.App
	Unique bool `json:"unique"`
}

func GetPluginApps(w http.ResponseWriter, r *http.Request) {
	a := rest.CTX(r).DB.AdminDB().Assets()

	db := rest.CTX(r).DB
	if db.Type() == database.PublicType || db.Type() == database.AppType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("Only logged in users can list available apps"))
		return
	}

	appmap := make(map[string]pluginApp)

	for pname, p := range a.Config.Plugins {
		for akey, app := range p.Apps {
			appid := pname + ":" + akey
			appmap[appid] = pluginApp{
				App:    plugins.App(appid, db.ID(), app),
				Unique: app.Unique != nil && *app.Unique,
			}
		}
	}

	rest.WriteJSON(w, r, appmap, nil)
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
