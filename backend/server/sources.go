package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
)

type Source struct {
	// Routes is nil if there is no backend routing
	Routes *chi.Mux

	// Create is nil if the source has no special creation handling
	Create http.Handler
}

// SourceManager handles sources
type SourceManager struct {
	Sources map[string]Source
	mux     *chi.Mux
	handler http.Handler
}

func stripRequestPrefix(h http.Handler, n int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Strip the first n elements from the path
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

// NewSourceManager generates a source manager
func NewSourceManager(a *assets.Assets, h http.Handler) (*SourceManager, error) {
	sources := make(map[string]Source)

	// Generate all handlers for the sources
	for sname, sv := range a.Config.SourceTypes {
		s := Source{}

		if sv.Routes != nil && len(*sv.Routes) > 0 {
			// The source has a backend component, which we now construct
			for r, uri := range *sv.Routes {
				fwd, err := NewReverseProxy(a.DataDir(), uri)
				if err != nil {
					return nil, err
				}
				if r == "create" {
					s.Create = fwd
				} else {
					if s.Routes == nil {
						s.Routes = chi.NewMux()
						s.Routes.NotFound(clearChiContext(h))
					}
					s.Routes.Handle(r, fwd)
				}
			}
		}
		sources[sname] = s
	}

	sm := &SourceManager{
		Sources: sources,
		mux:     chi.NewMux(),
		handler: h,
	}

	sm.mux.Post("/api/heedy/v1/source", sm.handleCreate)
	sm.mux.Mount("/api/heedy/v1/source/{sourceid}", http.HandlerFunc(sm.handleAPI))
	sm.mux.NotFound(sm.handler.ServeHTTP)

	return sm, nil
}

func (sm *SourceManager) handleCreate(w http.ResponseWriter, r *http.Request) {
	// Read the source in to find the type, and then see if we should forward the create request
	// or just handle it locally
	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(io.LimitReader(r.Body, *assets.Config().RequestBodyByteLimit))
	if err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("read_error: %s", err.Error()))
		return
	}
	r.Body.Close()

	var src database.Source

	if err = json.Unmarshal(data, &src); err != nil {
		WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("read_error: %s", err.Error()))
		return
	}
	if src.Type == nil {
		WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: must specify a type of source to create"))
		return
	}
	s, ok := sm.Sources[*src.Type]
	if !ok {
		WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: unrecognized source type"))
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

	s.Create.ServeHTTP(w, r)

}

func (sm *SourceManager) handleAPI(w http.ResponseWriter, r *http.Request) {
	// Get the source from the database, and find its type. Then, extract the scopes available for us
	// and set the X-Heedy-Scopes and X-Heedy-Source headers, and forward to the source API.

	// We need to clear the chi context if forwarding to the builtin REST API, because handleAPI was Mount-ed
	// which means that the context is relative to the mountpoint, whereas we want it to be the root context.
	sm.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, nil)))
}

func (sm *SourceManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v := r.Header.Get("X-Heedy-Overlay")
	if len(v) > 0 {
		oindex, err := strconv.Atoi(v)
		if err != nil {
			WriteJSONError(w, r, http.StatusBadRequest, errors.New("plugin_error: invalid X-Heedy-Overlay"))
			return
		}
		if oindex <= -1 {
			// The overlay is negative, meaning that we skip all source implementations
			sm.handler.ServeHTTP(w, r)
			return
		}

	}

	sm.mux.ServeHTTP(w, r)
}
