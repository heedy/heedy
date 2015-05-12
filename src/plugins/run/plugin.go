package run

/**
The plugin file specifies the interface needed to register ourselves with the
plugin registry when we're imported without side effects.
**/

import (
	"fmt"
	"log"
	"net/http"
	"plugins"
	"streamdb"
	"streamdb/config"

	"github.com/gorilla/mux"

	"plugins/rest"
	"plugins/webclient"
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
	http.Handle("/", r)

	return http.ListenAndServe(fmt.Sprintf(":%d", config.GetConfiguration().WebPort), nil)
}

func usage() {
	fmt.Println(`run: runs the HTTP and rest servers`)
}
