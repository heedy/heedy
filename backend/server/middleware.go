package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

// Middleware is a middleware that authenticates requests and generates a Context object containing
// the info necessary to complete the request
type Middleware struct {
	handler http.Handler

	auth *Auth

	// The auth system also allows special token-based access. This is specifically built
	// to support plugins. Each request that is forwarded through the plugin system
	// is first authenticated here, and given an auth token. Plugins can then make requests
	// with that auth token which will have the same permissions, and be linked to the original
	// request.
	sync.RWMutex
	activeRequests map[string]*Context
}

// NewMiddleware generates a new Auth middleware
func NewMiddleware(auth *Auth, h http.Handler) *Middleware {
	return &Middleware{
		auth:           auth,
		handler:        h,
		activeRequests: make(map[string]*Context),
	}
}

func (a *Middleware) serve(w http.ResponseWriter, r *http.Request, requestStart time.Time, c *Context) {
	a.Lock()
	a.activeRequests[c.Token] = c
	a.Unlock()
	a.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxtype, c)))
	a.Lock()
	delete(a.activeRequests, c.Token)
	a.Unlock()
	// Aaaand we're done here!
	c.Log.Debugf("%v", time.Since(requestStart))
}

// ServeHTTP - http.Handler implementation
func (a *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c *Context

	requestStart := time.Now()

	// First check if we are continuing an existing request
	authHeader := r.Header.Get("Authorization")

	// If the token comes from a "request handler", ie, a plugin processing an active request,
	// we don't generate a new context, but use an existing one.
	const handlerPrefix = "Handler "
	if len(authHeader) > len(handlerPrefix) && strings.EqualFold(handlerPrefix, authHeader[:len(handlerPrefix)]) {
		// The authorization header is of type Handler, this means that we might be in the middle of a request.
		// The request should be one of the active requests, so that we can continue it
		a.RLock()
		curRequest, ok := a.activeRequests[authHeader[len(handlerPrefix):]]
		a.RUnlock()
		if !ok {
			// The request was claiming to have a valid request token, but didn't!
			// This request ends *right here*
			requestLogger(r).Warn("Invalid handler token")

			// Sleep a second on invalid auth
			time.Sleep(time.Second)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid handler token"))
			return
		}

		// The request is active, we copy the relevant info from it to create a new context,
		// starting with a new logger
		logger := requestLogger(r).WithFields(logrus.Fields{
			"addr": curRequest.Log.Data["addr"], // addr is always the remote address
			"from": curRequest.Handler,          // from specifies the plugin that performed this request
			"id":   curRequest.ID,               // The request uses the original ID
			"auth": curRequest.Log.Data["auth"], // It also uses the same authentication
		})

		a.serve(w, r, requestStart, &Context{
			Log:   logger,
			ID:    curRequest.ID,
			DB:    curRequest.DB,
			Token: uuid.New().String(),
			From:  curRequest.Handler,
		})
		return

	}

	// The auth header does not have a handler token. This means that we are generating a context from scratch
	id := xid.New().String()
	c = &Context{
		Log:   requestLogger(r).WithField("id", id),
		ID:    id,
		Token: uuid.New().String(),
	}
	db, err := a.auth.Authenticate(r)
	if err != nil {
		// Authentication failed. This means that it was an illegal request, and we treat it as such
		time.Sleep(time.Second)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}
	c.DB = db
	c.Log = c.Log.WithField("auth", db.ID())

	a.serve(w, r, requestStart, c)

}
