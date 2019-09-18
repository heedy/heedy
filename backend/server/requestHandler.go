package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/plugins"
)

// RequestHandler is a middleware that authenticates requests and generates a Context object containing
// the info necessary to complete the request. It also handles generating and parsing the relevant X-Heedy headers
// that
type RequestHandler struct {
	auth    *Auth
	plugins *plugins.PluginManager

	// The auth system also allows special token-based access. This is specifically built
	// to support plugins. Each request that is forwarded through the plugin system
	// is first authenticated here, and given an auth token. Plugins can then make requests
	// with that auth token which will have the same permissions, and be linked to the original
	// request.
	sync.RWMutex
	activeRequests map[string]*rest.Context
}

// NewRequestHandler generates a new Auth middleware
func NewRequestHandler(auth *Auth,p *plugins.PluginManager) *RequestHandler {
	return &RequestHandler{
		auth:           auth,
		plugins:        p,
		activeRequests: make(map[string]*rest.Context),
	}
}

func (a *RequestHandler) serve(w http.ResponseWriter, r *http.Request, requestStart time.Time, c *rest.Context) {
	a.Lock()
	a.activeRequests[c.ID] = c
	a.Unlock()
	a.plugins.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), rest.HeedyContext, c)))
	a.Lock()
	delete(a.activeRequests, c.ID)
	a.Unlock()
	// Aaaand we're done here!
	c.Log.Debugf("%v", time.Since(requestStart))
}

// ServeHTTP - http.Handler implementation
func (a *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c *rest.Context

	requestStart := time.Now()

	logger := rest.RequestLogger(r)

	// First check if the request is coming from a plugin
	pluginKey := r.Header.Get("X-Heedy-Key")
	if len(pluginKey) > 0 {
		// There is a plugin key present, make sure it was given to one of the plugin processes
		proc, err := a.plugins.GetProcessByKey(pluginKey)
		if err!=nil {
			time.Sleep(time.Second)
			rest.WriteJSONError(w, r, http.StatusUnauthorized, errors.New("access_denied: invalid heedy plugin key"))
			return
		}

		logger = logger.WithField("plugin", proc.Plugin+"/"+proc.Exec)

		// Now check if it is a continuing request
		ID := r.Header.Get("X-Heedy-Id")
		if len(ID) > 0 {
			a.RLock()
			curRequest, ok := a.activeRequests[ID]
			a.RUnlock()
			if !ok {
				rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("plugin_error: invalid X-Heedy-Id"))
				return
			}

			// It is a continuing request! Let's pre-populate a bunch of values
			c = &rest.Context{
				RequestID: curRequest.RequestID,
				DB:        curRequest.DB,
			}
			logger = logger.WithField("addr", curRequest.Log.Data["addr"])

			// Remove the X-Heedy-Id

		} else {
			c = &rest.Context{
				RequestID: xid.New().String(),
				DB:        a.auth.DB,
			}

		}

		c.Plugin = proc.Plugin
		c.ID = uuid.New().String()

		// Now check if we are to update the context based on the X-Heedy headers
		authVal := r.Header.Get("X-Heedy-Auth")
		if len(authVal) > 0 && authVal != c.DB.ID() {
			c.DB, err = a.auth.As(authVal)
			if err != nil {
				rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("plugin_error: Could not auth as %s: %s", authVal, err.Error()))
				return
			}
		}

		// Finally, remove the X-Heedy-Key header, so that the plugin key isn't forwarded
		r.Header.Del("X-Heedy-Key")

		c.Log = logger.WithFields(logrus.Fields{
			"id":   c.RequestID,
			"auth": c.DB.ID(),
		})

	} else {

		// Make sure that there is no X-Heedy header in the request, because only plugins
		// are allowed to use those headers
		for header := range r.Header {
			if strings.HasPrefix(header, "X-Heedy") {
				rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("access_denied: X-Heedy headers are only permitted with a valid X-Heedy-Key"))
				return
			}
		}

		// No X-Heedy headers were found, this looks like a new request direct from the user
		id := xid.New().String()
		c = &rest.Context{
			Log:       logger.WithField("id", id),
			RequestID: id,
			ID:        uuid.New().String(),
		}

		db, err := a.auth.Authenticate(r)
		if err != nil {
			// Authentication failed. This means that it was an illegal request, and we treat it as such
			time.Sleep(time.Second)
			rest.WriteJSONError(w, r, http.StatusUnauthorized, fmt.Errorf("access_denied: %s", err.Error()))

			return
		}
		c.DB = db
		c.Log = c.Log.WithField("auth", db.ID())
	}

	// Set the appropriate X-Heedy Headers
	r.Header["X-Heedy-Auth"] = []string{c.DB.ID()}
	r.Header["X-Heedy-Id"] = []string{c.ID}
	r.Header["X-Heedy-Request"] = []string{c.RequestID}
	// Scopes?

	a.serve(w, r, requestStart, c)

}
