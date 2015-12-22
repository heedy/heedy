/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"connectordb"
	"net/http"
)

func JoinHandler(writer http.ResponseWriter, request *http.Request) {
	// TODO: Only show join page if allowed
	writer.WriteHeader(http.StatusOK)
	WWWJoin.Execute(writer, map[string]interface{}{"Version": connectordb.Version})
}
