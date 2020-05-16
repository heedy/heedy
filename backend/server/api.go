package server

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/heedy/heedy/api/golang/rest"
)

func APINotFound(w http.ResponseWriter, r *http.Request) {
	rest.WriteJSONError(w, r, http.StatusNotFound, rest.ErrNotFound)
}

// APIMux gives the REST API
func APIMux() (*chi.Mux, error) {

	apiMux := chi.NewMux()

	apiMux.Get("/events", EventWebsocket)
	apiMux.Post("/events", FireEvent)

	apiMux.Post("/users", CreateUser)
	apiMux.Get("/users", ListUsers)
	apiMux.Get("/users/{username}", ReadUser)
	apiMux.Patch("/users/{username}", UpdateUser)
	apiMux.Delete("/users/{username}", DeleteUser)

	apiMux.Post("/objects", CreateObject)
	apiMux.Get("/objects", ListObjects)
	apiMux.Get("/objects/{objectid}", ReadObject)
	apiMux.Patch("/objects/{objectid}", UpdateObject)
	apiMux.Delete("/objects/{objectid}", DeleteObject)

	apiMux.Post("/apps", CreateApp)
	apiMux.Get("/apps", ListApps)
	apiMux.Get("/apps/{appid}", ReadApp)
	apiMux.Patch("/apps/{appid}", UpdateApp)
	apiMux.Delete("/apps/{appid}", DeleteApp)

	apiMux.Get("/server/scope/{objecttype}", GetObjectScope)
	apiMux.Get("/server/scope", GetAppScope)
	apiMux.Get("/server/apps", GetPluginApps)
	apiMux.Get("/server/version", GetVersion)

	apiMux.Get("/server/admin", GetAdminUsers)
	apiMux.Post("/server/admin/{username}", AddAdminUser)
	apiMux.Delete("/server/admin/{username}", RemoveAdminUser)

	apiMux.Get("/server/updates", GetUpdates)
	apiMux.Delete("/server/updates", ClearUpdates)
	apiMux.Get("/server/updates/status", GetUpdateStatus)
	apiMux.Get("/server/updates/heedy.conf", GetConfigFile)
	apiMux.Post("/server/updates/heedy.conf", PostConfigFile)
	apiMux.Get("/server/updates/config", GetUConfig)
	apiMux.Patch("/server/updates/config", PatchUConfig)
	apiMux.Get("/server/updates/plugins", GetAllPlugins)
	apiMux.Post("/server/updates/plugins", PostPlugin)
	apiMux.Get("/server/updates/options", GetUpdateOptions)
	apiMux.Post("/server/updates/options", PostUpdateOptions)
	apiMux.Get("/server/updates/plugins/{pluginname}/README.md", GetPluginReadme)

	apiMux.NotFound(APINotFound)
	apiMux.MethodNotAllowed(APINotFound)
	return apiMux, nil
}
