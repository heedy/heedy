/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"connectordb"
	"connectordb/authoperator"
	"net/http"
	"server/webcore"

	"github.com/nu7hatch/gouuid"

	log "github.com/Sirupsen/logrus"
)

// These functions use the structures defined in templatehandlers

//WriteError writes the templated error page
func WriteError(logger *log.Entry, writer http.ResponseWriter, status int, err error, iserr bool) (int, string) {
	errmap := map[string]interface{}{
		"code": status,
		"msg":  err.Error(),
	}
	u, err2 := uuid.NewV4()
	if err2 != nil {
		logger.WithField("ref", "OSHIT").Errorln("Failed to generate error UUID: " + err2.Error())
		logger.WithField("ref", "OSHIT").Warningln("Original Error: " + err.Error())
		writer.WriteHeader(520)

		errmap["msg"] = "Failed to generate error UUID"
		errmap["ref"] = "OSHIT"
		return webcore.INFO, ""
	}
	errmap["ref"] = u.String()
	//Now that we have the error message, we log it and send the messages
	l := logger.WithFields(log.Fields{"ref": u.String(), "code": status})
	if iserr {
		l.Errorln(err.Error())
	} else {
		l.Warningln(err.Error())
	}

	writer.WriteHeader(status)
	AppError.Execute(writer, errmap)

	return webcore.INFO, ""
}

// LoggedIn404 sets up the 404 page for a logged in user. This is not an error page, since
// it is usually referring to a permissions error
func LoggedIn404(o *authoperator.AuthOperator, writer http.ResponseWriter, logger *log.Entry, oerr error) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	u, err := uuid.NewV4()
	if err != nil {
		return WriteError(logger, writer, 520, err, true)
	}

	td.Ref = u.String()
	td.Status = oerr.Error()

	App404.Execute(writer, td)

	return webcore.DEBUG, "404"
}

//NotFoundHandler handles all pages that were not found by writing the 404 templates
func NotFoundHandler(writer http.ResponseWriter, request *http.Request) {
	logger := webcore.GetRequestLogger(request, "404")
	writer.WriteHeader(http.StatusNotFound)

	//TODO: Make the LoggedIn404 work here

	// And a not-logged-in 404 page
	WWW404.Execute(writer, map[string]interface{}{"Version": connectordb.Version})

	//We give the overall 404 page
	logger.Debug("")
}
