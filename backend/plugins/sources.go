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

type Source struct {
	// Routes is nil if there is not backend routing
	Routes *chi.Mux

	// Create is nil if there is no special creation handling
	Create http.Handler
}
type SourceManager struct {
	A       *assets.Assets
	M       *run.Manager
	Sources map[string]Source
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

func NewSourceManager(a *assets.Assets, m *run.Manager, h http.Handler) (*SourceManager, error) {
	sources := make(map[string]Source)

	// The base handler is served from mounted handlers, so we need to make sure the chi context is cleared
	// so that it restarts the mux
	h = clearChiContext(h)

	// Generate all handlers for the sources that don't use any plugins
	for sname, sv := range a.Config.SourceTypes {
		s := Source{}

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
					logrus.Debugf("sources.%s: Forwarding %s -> %s", sname, r, uri)
					if r == "create" {
						s.Create = fwdstrip
					} else {
						run.Route(s.Routes, r, fwdstrip)
					}
				}
			}
		}
		sources[sname] = s
	}

	sm := &SourceManager{
		A:       a,
		M:       m,
		Sources: sources,
		mux:     chi.NewMux(),
		handler: h,
	}

	sm.mux.Post("/api/heedy/v1/sources", sm.handleCreate)
	// Since the Post is here, we must manually set the GET as valid and forward it
	// to the underlying api, otherwise we get a 405 error
	sm.mux.Get("/api/heedy/v1/sources", sm.handler.ServeHTTP)
	sm.mux.Mount("/api/heedy/v1/sources/{sourceid}", http.HandlerFunc(sm.handleAPI))
	sm.mux.NotFound(sm.handler.ServeHTTP)

	return sm, nil
}

func (sm *SourceManager) PreparePlugin(plugin string) error {
	// Generate the handlers for sources that explicitly use runs started by the given plugin
	for sname, sv := range sm.A.Config.SourceTypes {
		s := sm.Sources[sname]
		if sv.Routes != nil && len(*sv.Routes) > 0 {
			for r, uri := range *sv.Routes {
				pname, _, _ := run.GetPlugin("", uri)
				if pname == plugin {
					h, err := sm.M.GetHandler("", uri)
					if err != nil {
						return err
					}
					fwdstrip := stripRequestPrefix(h, 6)
					logrus.Debugf("sources.%s: Forwarding %s -> %s", sname, r, uri)
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

func (sm *SourceManager) handleCreate(w http.ResponseWriter, r *http.Request) {
	// Read the source in to find the type, and then see if we should forward the create request
	// or just handle it locally
	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(io.LimitReader(r.Body, *assets.Config().RequestBodyByteLimit))
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("read_error: %s", err.Error()))
		return
	}
	r.Body.Close()

	var src database.Source

	if err = json.Unmarshal(data, &src); err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("read_error: %s", err.Error()))
		return
	}
	if src.Type == nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: must specify a type of source to create"))
		return
	}
	s, ok := sm.Sources[*src.Type]
	if !ok {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: unrecognized source type"))
		return
	}

	// Looks like the request is valid. Recreate the request body so that it can be forwarded
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if s.Create == nil {
		// OK, so just forward the request to the standard API
		sm.handler.ServeHTTP(w, r)
		return
	}

	// There is a forward for this source type. First check if we have permission to create the source in the first place,
	// and then forward.
	err = rest.CTX(r).DB.CanCreateSource(&src)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusForbidden, err)
		return
	}

	s.Create.ServeHTTP(w, r)

}

func (sm *SourceManager) handleAPI(w http.ResponseWriter, r *http.Request) {
	// Get the source from the database, and find its type. Then, extract the scopes available for us
	// and set the X-Heedy-Scopes and X-Heedy-Source headers, and forward to the source API.
	ctx := rest.CTX(r)
	srcid := chi.URLParam(r, "sourceid")
	s, err := ctx.DB.ReadSource(srcid, nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusForbidden, err)
		return
	}
	lastModified := "null"
	if s.LastModified != nil {
		lastModified = (*s.LastModified).String()
	}
	r.Header["X-Heedy-Source"] = []string{srcid}
	r.Header["X-Heedy-Owner"] = []string{*s.Owner}
	r.Header["X-Heedy-Type"] = []string{*s.Type}
	r.Header["X-Heedy-Last-Modified"] = []string{lastModified}
	r.Header["X-Heedy-Access"] = s.Access.Scopes

	b, err := json.Marshal(s.Meta)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	r.Header["X-Heedy-Meta"] = []string{base64.StdEncoding.EncodeToString(b)}

	// Now get the correct source API
	ss, ok := sm.Sources[*s.Type]
	if ok {
		if ss.Routes != nil {
			ss.Routes.ServeHTTP(w, r)
			return
		}
	} else {
		ctx.Log.Warnf("Request is for an unrecognized source '%s'", *s.Type)
	}

	// We need to clear the chi context if forwarding to the builtin REST API, because handleAPI was Mount-ed
	// which means that the context is relative to the mountpoint, whereas we want it to be the root context.
	sm.handler.ServeHTTP(w, r)
}

func (sm *SourceManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v := r.Header.Get("X-Heedy-Overlay")
	if len(v) > 0 {
		if v == "none" {
			// No overlay, meaning that we skip all source implementations
			sm.handler.ServeHTTP(w, r)
			return
		}

	}

	sm.mux.ServeHTTP(w, r)
}
