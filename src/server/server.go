package server

import (
	"config"
	"connectordb"
	"fmt"
	"net/http"
	"server/restapi"
	"server/security"
	"server/webapp"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
)

//RunServer runs the ConnectorDB frontend server
func RunServer(c *config.Configuration) error {
	//Connect using the configuration
	db, err := connectordb.Open(c.Options())
	if err != nil {
		return err
	}

	r := mux.NewRouter()

	webapp.Setup(r, db)

	//The rest api has its own versioned url
	s := r.PathPrefix("/api/v1").Subrouter()
	restapi.Router(db, s)

	//All else goes to the web server
	http.Handle("/", security.NewSecurityBuilder(r).
		IncludeSecureHeaders().
		Build())

	//Run the dbwriter
	go db.RunWriter()

	log.Infof("Running ConnectorDB v%s at %s:%d", connectordb.Version, c.Hostname, c.Port)

	return http.ListenAndServe(fmt.Sprintf("%s:%d", c.Hostname, c.Port), nil)
}
