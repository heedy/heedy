package rest

import (
	"io"
	"net/http"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"

	"github.com/sirupsen/logrus"
)

// notABasicType is there because the context package doesn't like basic indices, so we use a constant that is
// used to get the context from requests
type notABasicType uint8

const (
	// HeedyContext is the context entry used to get the auth context from the request context
	HeedyContext notABasicType = iota
)

// CTX gets the heedy request context from an http.Request
func CTX(r *http.Request) *Context {
	hc := r.Context().Value(HeedyContext)
	if hc == nil {
		return nil
	}
	return hc.(*Context)
}

type Requester interface {
	Request(c *Context, method, path string, body interface{}, header map[string]string) (io.Reader, error)
}

// A Context is generated for all requests, and holds all the info necessary for completing it.
// This object can be extracted from a request with the CTX function.
type Context struct {
	Requester

	Log       *logrus.Entry  // The request's logger
	DB        database.DB    // The authenticated database object
	Events    events.Handler // The event handler (must be set here for plugins)
	RequestID string         // The ID of the original query to the API

	ID     string // The ID sent to plugins in X-Heedy-ID header, and is used for all internal requests
	Plugin string // The plugin that sent the request
}
