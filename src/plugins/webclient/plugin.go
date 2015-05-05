package webclient

/**
The plugin file specifies the interface needed to register ourselves with the
plugin registry when we're imported without side effects.
**/

import (
	"streamdb"
	"plugins"
	"fmt"
	"log"
	"streamdb/config"
	"net/http"
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
	http.Handle("/", r)

	return http.ListenAndServe(fmt.Sprintf(":%d", config.GetConfiguration().WebPort), nil)
}

func usage() {
	fmt.Println(`web: runs the HTTP server users can interact with.`)
}
