package webclient

/**
The plugin file specifies the interface needed to register ourselves with the
plugin registry when we're imported without side effects.
**/

import (
	"connectordb/config"
	"connectordb/plugins"
	"connectordb/security"
	"connectordb/streamdb"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

func init() {
	// do some sweet plugin registration!
	plugins.Register("web", usage, exec)
}

func exec(db *streamdb.Database, args []string) error {
	log.Printf("Starting Server on port %d", config.GetConfiguration().WebPort)
	r := mux.NewRouter()
	Setup(r, db)
	http.Handle("/", security.SecurityHeaderHandler(r))

	return http.ListenAndServe(fmt.Sprintf(":%d", config.GetConfiguration().WebPort), nil)
}

func usage() {
	fmt.Println(`web: runs the HTTP server users can interact with.`)
}
