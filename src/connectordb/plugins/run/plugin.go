package run

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

	"connectordb/plugins/rest"
	"connectordb/plugins/webclient"
)

func init() {
	// do some sweet plugin registration!
	plugins.Register("run", usage, exec)
}

func exec(db *streamdb.Database, args []string) error {
	log.Printf("Starting Server on port %d", config.GetConfiguration().WebPort)
	r := mux.NewRouter()
	webclient.Setup(r, db)

	// handle the api at its versioned url
	s := r.PathPrefix("/api/v1").Subrouter()
	rest.Router(db, s)

	// all else goes to the webserver
	http.Handle("/", security.SecurityHeaderHandler(r))

	go db.RunWriter()

	return http.ListenAndServe(fmt.Sprintf(":%d", config.GetConfiguration().WebPort), nil)
}

func usage() {
	fmt.Println(`run: Runs the full ConnectorDB system`)
}
