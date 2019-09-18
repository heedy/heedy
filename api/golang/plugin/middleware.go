package plugin

import (
	"context"
	"net/http"

	"github.com/heedy/heedy/api/golang/rest"
)

// Middleware constructs a Heedy server context for an http handler, allowing
// to create API handlers compatible with the heedy builtin server. That is,
// a plugin that uses the middleware can in the future be embedded in Heedy without any changes.
type Middleware struct {
	P *Plugin
	H http.Handler
}

func NewMiddleware(p *Plugin, h http.Handler) *Middleware {
	return &Middleware{
		P: p,
		H: h,
	}
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := rest.RequestLogger(r)
	logger = logger.WithField("exec", m.P.Meta.Plugin+"/"+m.P.Meta.Exec)
	// Read out the header values necessary to generate a context

	c := rest.Context{
		Log:       logger,
		RequestID: r.Header.Get("X-Heedy-RequestID"),
		ID:        r.Header.Get("X-Heedy-ID"),
		DB:        m.P.As(r.Header.Get("X-Heedy-Auth")),
	}

	r = r.WithContext(context.WithValue(r.Context(), rest.HeedyContext, &c))
	m.H.ServeHTTP(w, r)
}
