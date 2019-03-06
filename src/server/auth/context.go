package auth

import (
	"github.com/connectordb/connectordb/src/database"

	log "github.com/sirupsen/logrus"
)

// A Context is generated for all requests, and holds all the info necessary for completing it.
// This object is injected into each request's context under the key Ctx
type Context struct {
	Log     *log.Entry  // The request's logger
	DB      database.DB // The authenticated database object
	ID      string      // The request ID
	Token   string      // An auth token to allow continuing requests that were forwarded to plugins
	From    string      // The handler that sent the request. Is empty when request is straight from user
	Handler string      // The handler that is processing the request. Is empty when starting - set by the plugin system
}
