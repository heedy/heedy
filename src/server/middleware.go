package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"runtime/debug"

	"github.com/connectordb/connectordb/src/database"

	log "github.com/sirupsen/logrus"
)

// context Key, because go gives some bullshit about "no strings allowed"
type cK string

// An AuthHandler acts as a middleware which authenticates users of ConnectorDB.
// If there is no authentication in the request, it adds public db to context.
// If there is auth, it denies access if unauthenticated, and adds logged in db to context if valid auth.
type AuthHandler struct {
	db      *database.AdminDB
	handler http.Handler
}

// NewAuthHandler generates a new AuthHandler
func NewAuthHandler(handler http.Handler, db *database.AdminDB) *AuthHandler {
	return &AuthHandler{
		db:      db,
		handler: handler,
	}
}

// ServeHTTP - http.Handler implementation
func (a *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Basic auth is first
	_, _, ok := r.BasicAuth()
	if ok {

	}
	a.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), cK("cdb"), a.db)))
}

// LoggingHandler handles logging - it returns a middleware that prints out relevant request info
type LoggingHandler struct {
	handler http.Handler
}

// https://www.reddit.com/r/golang/comments/7p35s4/how_do_i_get_the_response_status_for_my_middleware/
type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// NewLoggingHandler returns a middleware for logging
func NewLoggingHandler(handler http.Handler) *LoggingHandler {
	return &LoggingHandler{
		handler: handler,
	}
}

func (l *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fields := log.Fields{"addr": r.RemoteAddr, "path": r.URL.Path, "method": r.Method}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		fields["realip"] = realIP
	}
	logger := log.WithFields(fields)

	sw := statusWriter{ResponseWriter: w}

	defer func() {
		if p := recover(); p != nil {
			req, err := httputil.DumpRequest(r, true)
			if err != nil {
				req = []byte(fmt.Sprintf("Error dumping Request: %s", err))
			}
			sb := debug.Stack()

			// We might want to pass the panic on, but for now, whatever
			logger.Logf(log.PanicLevel, "%s\n\n%s\n\n%s\n", p, string(req), string(sb))
		}
	}()

	l.handler.ServeHTTP(&sw, r.WithContext(context.WithValue(r.Context(), cK("log"), logger)))

	logger.Infof("%d", sw.status)
}
