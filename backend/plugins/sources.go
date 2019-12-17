package plugins

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/sirupsen/logrus"
)

type Object struct {
	// Routes is nil if there is not backend routing
	Routes *chi.Mux

	// Create is nil if there is no special creation handling
	Create http.Handler
}
type ObjectManager struct {
	A       *assets.Assets
	M       *run.Manager
	Objects map[string]Object
	mux     *chi.Mux
	handler http.Handler
}

func stripRequestPrefix(h http.Handler, n int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Strip the first n elements from the path
		if r.URL.Path[0] == '/' {
			r.URL.Path = r.URL.Path[1:]
		}
		s := strings.SplitN(r.URL.Path, "/", n)
		if len(s) < n {
			r.URL.Path = "/"
		} else {
			r.URL.Path = "/" + s[len(s)-1]
		}

		h.ServeHTTP(w, r)
	}
}

// chi modifies its context for subrouters (ie: when Mount is used, it scopes the context)
// We want to clear the context scoping
func clearChiContext(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, nil)))
	}
}

func NewObjectManager(a *assets.Assets, m *run.Manager, h http.Handler) (*ObjectManager, error) {
	objects := make(map[string]Object)

	// The base handler is served from mounted handlers, so we need to make sure the chi context is cleared
	// so that it restarts the mux
	h = clearChiContext(h)

	// Generate all handlers for the objects that don't use any plugins
	for sname, sv := range a.Config.ObjectTypes {
		s := Object{}

		if sv.Routes != nil && len(*sv.Routes) > 0 {
			for r, uri := range *sv.Routes {
				if r != "create" && s.Routes == nil {
					s.Routes = chi.NewMux()
					s.Routes.NotFound(h.ServeHTTP)
				}
				plugin, _, _ := run.GetPlugin("", uri)
				if plugin == "" {
					h, err := m.GetHandler("", uri)
					if err != nil {
						return nil, err
					}
					fwdstrip := stripRequestPrefix(h, 6)
					logrus.Debugf("objects.%s: Forwarding %s -> %s", sname, r, uri)
					if r == "create" {
						s.Create = fwdstrip
					} else {
						run.Route(s.Routes, r, fwdstrip)
					}
				}
			}
		}
		objects[sname] = s
	}

	sm := &ObjectManager{
		A:       a,
		M:       m,
		Objects: objects,
		mux:     chi.NewMux(),
		handler: h,
	}

	sm.mux.Post("/api/heedy/v1/objects", sm.handleCreate)
	// Since the Post is here, we must manually set the GET as valid and forward it
	// to the underlying api, otherwise we get a 405 error
	sm.mux.Get("/api/heedy/v1/objects", sm.handler.ServeHTTP)
	sm.mux.Mount("/api/heedy/v1/objects/{objectid}", http.HandlerFunc(sm.handleAPI))
	sm.mux.NotFound(sm.handler.ServeHTTP)

	return sm, nil
}

func (sm *ObjectManager) PreparePlugin(plugin string) error {
	// Generate the handlers for objects that explicitly use runs started by the given plugin
	for sname, sv := range sm.A.Config.ObjectTypes {
		s := sm.Objects[sname]
		if sv.Routes != nil && len(*sv.Routes) > 0 {
			for r, uri := range *sv.Routes {
				pname, _, _ := run.GetPlugin("", uri)
				if pname == plugin {
					h, err := sm.M.GetHandler("", uri)
					if err != nil {
						return err
					}
					fwdstrip := stripRequestPrefix(h, 6)
					logrus.Debugf("objects.%s: Forwarding %s -> %s", sname, r, uri)
					if r == "create" {
						s.Create = fwdstrip
					} else {
						err := run.Route(s.Routes, r, fwdstrip)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (sm *ObjectManager) handleCreate(w http.ResponseWriter, r *http.Request) {
	// Read the object in to find the type, and then see if we should forward the create request
	// or just handle it locally
	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(io.LimitReader(r.Body, *assets.Config().RequestBodyByteLimit))
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("read_error: %s", err.Error()))
		return
	}
	r.Body.Close()

	var src database.Object

	if err = json.Unmarshal(data, &src); err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("read_error: %s", err.Error()))
		return
	}
	if src.Type == nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: must specify a type of object to create"))
		return
	}
	s, ok := sm.Objects[*src.Type]
	if !ok {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: unrecognized object type"))
		return
	}

	// Looks like the request is valid. Recreate the request body so that it can be forwarded
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if s.Create == nil {
		// OK, so just forward the request to the standard API
		sm.handler.ServeHTTP(w, r)
		return
	}

	// There is a forward for this object type. First check if we have permission to create the object in the first place,
	// and then forward.
	err = rest.CTX(r).DB.CanCreateObject(&src)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusForbidden, err)
		return
	}

	s.Create.ServeHTTP(w, r)

}

func (sm *ObjectManager) handleAPI(w http.ResponseWriter, r *http.Request) {
	// Get the object from the database, and find its type. Then, extract the scopes available for us
	// and set the X-Heedy-Scope and X-Heedy-Object headers, and forward to the object API.
	ctx := rest.CTX(r)
	srcid := chi.URLParam(r, "objectid")
	s, err := ctx.DB.ReadObject(srcid, nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusForbidden, err)
		return
	}
	lastModified := "null"
	if s.LastModified != nil {
		lastModified = (*s.LastModified).String()
	}
	r.Header["X-Heedy-Object"] = []string{srcid}
	r.Header["X-Heedy-Owner"] = []string{*s.Owner}
	r.Header["X-Heedy-Type"] = []string{*s.Type}
	r.Header["X-Heedy-Last-Modified"] = []string{lastModified}
	r.Header["X-Heedy-Access"] = s.Access.Scope

	b, err := json.Marshal(s.Meta)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	r.Header["X-Heedy-Meta"] = []string{base64.StdEncoding.EncodeToString(b)}

	// Now get the correct object API
	ss, ok := sm.Objects[*s.Type]
	if ok {
		if ss.Routes != nil {
			ss.Routes.ServeHTTP(w, r)
			return
		}
	} else {
		ctx.Log.Warnf("Request is for an unrecognized object '%s'", *s.Type)
	}

	// We need to clear the chi context if forwarding to the builtin REST API, because handleAPI was Mount-ed
	// which means that the context is relative to the mountpoint, whereas we want it to be the root context.
	sm.handler.ServeHTTP(w, r)
}

func (sm *ObjectManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v := r.Header.Get("X-Heedy-Overlay")
	if len(v) > 0 {
		if v == "none" {
			// No overlay, meaning that we skip all object implementations
			sm.handler.ServeHTTP(w, r)
			return
		}

	}

	sm.mux.ServeHTTP(w, r)
}
