/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package query

import (
	"connectordb"
	"connectordb/authoperator"
	"connectordb/query"
	"fmt"
	"net/http"
	"server/restapi/restcore"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
)

//GenerateDataset allows to generate a dataset of multiple streams at once to simplify analysis of data
func GenerateDataset(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	var datasetquery query.DatasetQuery
	err := restcore.UnmarshalRequest(request, &datasetquery)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	dr, err := datasetquery.Run(o)
	return restcore.WriteJSONResult(writer, dr, logger, err)
}

//MergeStreams allows to generate a dataset of multiple streams at once to simplify analysis of data
func MergeStreams(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	var mergequery []*query.StreamQuery
	err := restcore.UnmarshalRequest(request, &mergequery)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	dr, err := query.Merge(o, mergequery)
	lvl, _ := restcore.WriteJSONResult(writer, dr, logger, err)
	return lvl, fmt.Sprintf("Merging %d streams", len(mergequery))
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *connectordb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.HandleFunc("/dataset", restcore.Authenticator(GenerateDataset, db)).Methods("POST")
	prefix.HandleFunc("/merge", restcore.Authenticator(MergeStreams, db)).Methods("POST")

	return prefix
}
