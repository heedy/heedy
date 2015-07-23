package rest

import (
	"connectordb/streamdb"
	"errors"

	"net/http"
	"sync/atomic"

	"github.com/gorilla/mux"

	"connectordb/plugins/rest/restcore"
	"connectordb/plugins/rest/restd"
)

//NotFoundHandler when a path is not found, return a 404 with path not recognized message
func NotFoundHandler(writer http.ResponseWriter, request *http.Request) {
	atomic.AddUint32(&restcore.StatsQueries, 1)
	logger := restcore.GetRequestLogger(request, "404")
	restcore.WriteError(writer, logger, http.StatusNotFound, errors.New("This path is not recognized"), false)
}

//OptionsHandler on OPTIONS to allow cross-site XMLHTTPRequest, allow access control origin
func OptionsHandler(writer http.ResponseWriter, request *http.Request) {
	atomic.AddUint32(&restcore.StatsQueries, 1)
	restcore.GetRequestLogger(request, "OPTIONS").Debug()
	restcore.WriteAccessControlHeaders(writer)
	writer.WriteHeader(http.StatusOK)
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	prefix.Methods("OPTIONS").Handler(http.HandlerFunc(OptionsHandler))

	// The websocket is run straight from here
	prefix.HandleFunc("/", restcore.Authenticator(RunWebsocket, db)).Headers("Upgrade", "websocket").Methods("GET")

	//The 'd' prefix corresponds to data
	restd.Router(db, prefix.PathPrefix("/d").Subrouter())

	go restcore.RunStats()
	go restcore.RunQueryTimers()

	return prefix
}
