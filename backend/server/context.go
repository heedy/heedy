package server

import (
	"github.com/heedy/heedy/backend/database"

	log "github.com/sirupsen/logrus"
)

// A Context is generated for all requests, and holds all the info necessary for completing it.
// This object can be extracted from a request with the CTX function.
type Context struct {
	Log *log.Entry  // The request's logger
	DB  database.DB // The authenticated database object
	ID  string      // The ID of the original query

	RequestID string   // The ID sent to plugins in X-Heedy-ID header
	Plugin    string   // The plugin that sent the request
	Scopes    []string // The scopes to enable at this route
}
