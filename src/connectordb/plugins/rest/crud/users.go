package crud

import (
	"connectordb/plugins/rest/restcore"
	"connectordb/streamdb/operator"

	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

//ListUsers lists the users that the given operator can see
func ListUsers(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
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
func CreateUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	usrname := mux.Vars(request)["user"]
	var a userCreator
	err := restcore.UnmarshalRequest(request, &a)
	err = restcore.ValidName(usrname, err)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	if err = o.CreateUser(usrname, a.Email, a.Password); err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}

	//Creating user is an info level event
	logger.Infoln()

	return ReadUser(o, writer, request, logger)
}

//ReadUser reads the given user
func ReadUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	usrname := mux.Vars(request)["user"]

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	u, err := o.ReadUser(usrname)

	if err == nil {
		u.Password = ""
	}

	return restcore.JSONWriter(writer, u, logger, err)
}

//UpdateUser updates the metadata for existing user from a REST API request
func UpdateUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	usrname := mux.Vars(request)["user"]

	u, err := o.ReadUser(usrname)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}

	modusr := *u
	err = restcore.UnmarshalRequest(request, &modusr)
	err = restcore.ValidName(modusr.Name, err)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	//We use a special procedure for upgrading the password
	if modusr.Password != u.Password {
		modusr.SetNewPassword(modusr.Password)
	}
	if err = o.UpdateUser(&modusr); err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}
	return restcore.JSONWriter(writer, modusr, logger, err)
}

//DeleteUser deletes existing user from a REST API request
func DeleteUser(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	usrname := mux.Vars(request)["user"]
	err := o.DeleteUser(usrname)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}
	logger.Infoln()
	return restcore.OK(writer)
}
