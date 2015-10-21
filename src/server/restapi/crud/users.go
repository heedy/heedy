package crud

import (
	"server/restapi/restcore"
	"connectordb/operator"

	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

//ListUsers lists the users that the given operator can see
func ListUsers(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	u, err := o.ReadAllUsers()
	if err != nil {
		for i := 0; i < len(u); i++ {
			u[i].Password = ""
		}
	}
	return restcore.JSONWriter(writer, u, logger, err)
}

type userCreator struct {
	Email    string
	Password string
}

//CreateUser creates a new user from a REST API request
func CreateUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]
	var a userCreator
	err := restcore.UnmarshalRequest(request, &a)
	err = restcore.ValidName(usrname, err)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)

	}

	if err = o.CreateUser(usrname, a.Email, a.Password); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)

	}
	ReadUser(o, writer, request, logger)
	return restcore.INFO, ""
}

//ReadUser reads the given user
func ReadUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	u, err := o.ReadUser(usrname)

	if err == nil {
		u.Password = ""
	}

	return restcore.JSONWriter(writer, u, logger, err)
}

//UpdateUser updates the metadata for existing user from a REST API request
func UpdateUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]

	u, err := o.ReadUser(usrname)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}

	modusr := *u
	err = restcore.UnmarshalRequest(request, &modusr)
	err = restcore.ValidName(modusr.Name, err)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	//We use a special procedure for upgrading the password
	if modusr.Password != u.Password {
		modusr.SetNewPassword(modusr.Password)
	}
	if err = o.UpdateUser(&modusr); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	return restcore.JSONWriter(writer, modusr, logger, err)
}

//DeleteUser deletes existing user from a REST API request
func DeleteUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]
	err := o.DeleteUser(usrname)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	restcore.OK(writer)
	return restcore.INFO, ""
}
