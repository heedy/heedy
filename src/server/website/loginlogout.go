/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package website

import (
	"connectordb/authoperator"
	"net/http"
	"server/webcore"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Login handles login to the system without the api call (direct web interface)
func Login(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	if o.Name() == "nobody" {
		// If the operator is a nobody, backtrack to the login page
		return -1, ""
	}
	webcore.CreateSessionCookie(o, writer, request)
	http.Redirect(writer, request, "/", http.StatusFound)

	return webcore.DEBUG, ""
}

// LogoutHandler handles log out of the system without an api call (direct web interface)
func LogoutHandler(writer http.ResponseWriter, request *http.Request) {
	tstart := time.Now()
	logger := webcore.GetRequestLogger(request, "logout")

	//We don't need the "op" here
	delete(logger.Data, "op")

	webcore.CreateSessionCookie(nil, writer, request) //nil operator deletes the cookie
	http.Redirect(writer, request, "/", http.StatusFound)

	webcore.LogRequest(logger, webcore.DEBUG, "", time.Since(tstart))
}
