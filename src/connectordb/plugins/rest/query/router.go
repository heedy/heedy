package query

import (
	"connectordb/plugins/rest/restcore"
	"connectordb/streamdb"
	"connectordb/streamdb/operator"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
)

//GenerateDataset allows to generate a dataset of multiple streams at once to simplify analysis of data
func GenerateDataset(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	return restcore.WriteError(writer, logger, http.StatusNotImplemented, errors.New("This function is under construction"), false)
}

//MergeStreams allows to generate a dataset of multiple streams at once to simplify analysis of data
func MergeStreams(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	return restcore.WriteError(writer, logger, http.StatusNotImplemented, errors.New("This function is under construction"), false)
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.HandleFunc("/dataset", restcore.Authenticator(GenerateDataset, db)).Methods("GET")
	prefix.HandleFunc("/merge", restcore.Authenticator(MergeStreams, db)).Methods("GET")

	return prefix
}
