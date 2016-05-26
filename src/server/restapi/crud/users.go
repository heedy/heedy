/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package crud

import (
	"connectordb/authoperator"
	"connectordb/users"
	"server/restapi/restcore"
	"server/webcore"

	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

//ListUsers lists the users that the given operator can see
func ListUsers(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	u, err := o.ReadAllUsersToMap()
	return restcore.JSONWriter(writer, u, logger, err)
}

//CreateUser creates a new user from a REST API request
func CreateUser(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]

	err := restcore.ValidName(usrname, nil)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	var um users.UserMaker
	err = restcore.UnmarshalRequest(request, &um)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	um.Name = usrname
	if err = o.CreateUser(&um); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)

	}
	ReadUser(o, writer, request, logger)
	return webcore.INFO, ""
}

//ReadUser reads the given user
func ReadUser(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	u, err := o.ReadUserToMap(usrname)

	return restcore.JSONWriter(writer, u, logger, err)
}

//UpdateUser updates the metadata for existing user from a REST API request
func UpdateUser(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]

	var modusr map[string]interface{}
	err := restcore.UnmarshalRequest(request, &modusr)

	if err = o.UpdateUser(usrname, modusr); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	u, err := o.ReadUserToMap(usrname)
	return restcore.JSONWriter(writer, u, logger, err)
}

//DeleteUser deletes existing user from a REST API request
func DeleteUser(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]
	err := o.DeleteUser(usrname)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	restcore.OK(writer)
	return webcore.INFO, ""
}
