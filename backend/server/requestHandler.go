package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/heedy/heedy/backend/plugin"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

// notABasicType is there because the context package doesn't like basic indices, so we use a constant that is
// used to get the context from requests
type notABasicType uint8

const (
	// ctx is the context entry used to get the auth context from the request context
	ctxtype notABasicType = iota
)

// CTX gets the heedy request context from an http.Request
func CTX(r *http.Request) *Context {
	return r.Context().Value(ctxtype).(*Context)
}

// requestLogger generates a basic logger that holds relevant request info
func requestLogger(r *http.Request) *logrus.Entry {
	fields := logrus.Fields{"addr": r.RemoteAddr, "path": r.URL.Path, "method": r.Method}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		fields["realip"] = realIP
	}
	return logrus.WithFields(fields)
}

// RequestHandler is a middleware that authenticates requests and generates a Context object containing
// the info necessary to complete the request. It also handles generating and parsing the relevant X-Heedy headers
// that
type RequestHandler struct {
	auth    *Auth
	plugins *plugin.Manager

	// The auth system also allows special token-based access. This is specifically built
	// to support plugins. Each request that is forwarded through the plugin system
	// is first authenticated here, and given an auth token. Plugins can then make requests
	// with that auth token which will have the same permissions, and be linked to the original
	// request.
	sync.RWMutex
	activeRequests map[string]*Context
}

// NewRequestHandler generates a new Auth middleware
func NewRequestHandler(auth *Auth, m *plugin.Manager) *RequestHandler {
	return &RequestHandler{
		auth:           auth,
		plugins:        m,
		activeRequests: make(map[string]*Context),
	}
}

func (a *RequestHandler) serve(w http.ResponseWriter, r *http.Request, requestStart time.Time, c *Context) {
	a.Lock()
	a.activeRequests[c.RequestID] = c
	a.Unlock()
	a.plugins.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxtype, c)))
	a.Lock()
	delete(a.activeRequests, c.RequestID)
	a.Unlock()
	// Aaaand we're done here!
	c.Log.Debugf("%v", time.Since(requestStart))
}

// ServeHTTP - http.Handler implementation
func (a *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c *Context
	var err error

	requestStart := time.Now()

	logger := requestLogger(r)

	// First check if the request is coming from a plugin
	pluginKey := r.Header.Get("X-Heedy-Key")
	if len(pluginKey) > 0 {
		// There is a plugin key present
		proc, ok := a.plugins.Exec.Processes[pluginKey]
		if !ok {
			logger.Warn("Request has invalid X-Heedy-Key")
			time.Sleep(time.Second)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "access_denied", "error_description": "Invalid heedy plugin key"}`))
			return
		}

		logger = logger.WithField("plugin", proc.Plugin+"/"+proc.Exec)

		// Now check if it is a continuing request
		requestID := r.Header.Get("X-Heedy-Id")
		if len(requestID) > 0 {
			a.RLock()
			curRequest, ok := a.activeRequests[requestID]
			a.RUnlock()
			if !ok {
				logger.Warn("Request has invalid X-Heedy-Id")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "plugin_error", "error_description": "Invalid request ID"}`))
				return
			}

			// It is a continuing request! Let's pre-populate a bunch of values
			c = &Context{
				ID:     curRequest.ID,
				DB:     curRequest.DB,
				Scopes: curRequest.Scopes,
			}
			logger = logger.WithField("addr", curRequest.Log.Data["addr"])

			// Remove the X-Heedy-Id

		} else {
			c = &Context{
				ID:     xid.New().String(),
				DB:     a.auth.DB,
				Scopes: []string{"*"},
			}

		}

		c.Plugin = proc.Plugin + "/" + proc.Exec
		c.RequestID = uuid.New().String()

		// Now check if we are to update the context based on the X-Heedy headers
		authVal := r.Header.Get("X-Heedy-Auth")
		if len(authVal) > 0 && authVal != c.DB.ID() {
			c.DB, err = a.auth.As(authVal)
			if err != nil {
				logger.Warnf("Could not auth as %s: %s", authVal, err.Error())
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "plugin_error", "error_description": "Given auth not valid"}`))
				return
			}
		}

		// Finally, remove the X-Heedy-Key header, so that the plugin key isn't forwarded
		r.Header.Del("X-Heedy-Key")

		c.Log = logger.WithFields(logrus.Fields{
			"id":   c.ID,
			"auth": c.DB.ID(),
		})

	} else {

		// Make sure that there is no X-Heedy header in the request, because only plugins
		// are allowed to use those headers
		for header := range r.Header {
			if strings.HasPrefix(header, "X-Heedy") {
				logger.Warn("Request contains unauthorized X-Heedy headers")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "header_forbidden", "error_description": "X-Heedy headers are only permitted when X-Heedy-Key is set to a valid plugin key"}`))
				return
			}
		}

		// No X-Heedy headers were found, this looks like a new request direct from the user
		id := xid.New().String()
		c = &Context{
			Log:       logger.WithField("id", id),
			ID:        id,
			RequestID: uuid.New().String(),
		}

		db, err := a.auth.Authenticate(r)
		if err != nil {
			// Authentication failed. This means that it was an illegal request, and we treat it as such
			logger.Warnf("Authentication Failed: %s", err.Error())
			time.Sleep(time.Second)

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "access_denied", "error_description": "Could not authenticate"}`))
			return
		}
		c.DB = db
		c.Log = c.Log.WithField("auth", db.ID())
	}

	// Set the appropriate X-Heedy Headers
	r.Header["X-Heedy-Auth"] = []string{c.DB.ID()}
	r.Header["X-Heedy-Id"] = []string{c.RequestID}

	// Scopes?

	a.serve(w, r, requestStart, c)

}
