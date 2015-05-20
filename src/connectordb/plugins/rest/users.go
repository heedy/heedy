package rest

import (
	"connectordb/streamdb/operator"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

//ListUsers lists the users that the given operator can see
func ListUsers(o operator.Operator, writer http.ResponseWriter, request *http.Request) error {
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "ListUsers"})
	logger.Debugln()
	u, err := o.ReadAllUsers()
	if err != nil {
		for i := 0; i < len(u); i++ {
			u[i].Password = ""
		}
	}
	return JSONWriter(writer, u, logger, err)
}

type userCreator struct {
	Email    string
	Password string
}

//CreateUser creates a new user from a REST API request
func CreateUser(o operator.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := mux.Vars(request)["user"]
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "CreateUser", "arg": usrname})
	logger.Infoln()
	var a userCreator
	err := UnmarshalRequest(request, &a)
	err = ValidName(usrname, err)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}

	if err = o.CreateUser(usrname, a.Email, a.Password); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}

	return ReadUser(o, writer, request)
}

//ReadUser reads the given user
func ReadUser(o operator.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := mux.Vars(request)["user"]

	if err := BadQ(o, writer, request, usrname); err != nil {
		return err
	}

	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "ReadUser", "arg": usrname})
	logger.Debugln()
	u, err := o.ReadUser(usrname)

	if err == nil {
		u.Password = ""
	}

	return JSONWriter(writer, u, logger, err)
}

//UpdateUser updates the metadata for existing user from a REST API request
func UpdateUser(o operator.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := mux.Vars(request)["user"]
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "UpdateUser", "arg": usrname})
	logger.Infoln()
	u, err := o.ReadUser(usrname)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}

	modusr := *u
	err = UnmarshalRequest(request, &modusr)
	err = ValidName(modusr.Name, err)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}

	//We use a special procedure for upgrading the password
	if modusr.Password != u.Password {
		modusr.SetNewPassword(modusr.Password)
	}
	if err = o.UpdateUser(&modusr); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	return JSONWriter(writer, modusr, logger, err)
}

//DeleteUser deletes existing user from a REST API request
func DeleteUser(o operator.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := mux.Vars(request)["user"]
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "DeleteUser", "arg": usrname})
	logger.Infoln()
	err := o.DeleteUser(usrname)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	return OK(writer)
}
