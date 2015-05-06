package rest

import (
	"log"
	"net/http"
	"streamdb"
	"strings"

	"github.com/gorilla/mux"
)

//GetUser runs the GET operation routing for REST
func GetUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := strings.ToLower(mux.Vars(request)["user"])

	//there can be certain commands in place of a username - those represent invalid user names
	switch usrname {
	default:
		return ReadUser(o, writer, request)
	case "ls":
		return ListUsers(o, writer, request)
	case "this":
		return GetThis(o, writer, request)
	}

}

//ListUsers lists the users that the given operator can see
func ListUsers(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	log.Printf("%s: List Users\n", o.Name())
	u, err := o.ReadAllUsers()
	if err != nil {
		for i := 0; i < len(u); i++ {
			u[i].Password = ""
		}
	}
	return JSONWriter(writer, u, err)
}

type userCreator struct {
	Email    string
	Password string
}

//CreateUser creates a new user from a REST API request
func CreateUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := strings.ToLower(mux.Vars(request)["user"])
	log.Printf("%s: Create User %s\n", o.Name(), usrname)
	var a userCreator
	err := UnmarshalRequest(request, &a)
	err = ValidName(usrname, err)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return err
	}

	if err = o.CreateUser(usrname, a.Email, a.Password); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}

	return ReadUser(o, writer, request)
}

//ReadUser reads the given user
func ReadUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := strings.ToLower(mux.Vars(request)["user"])
	log.Printf("%s: Read User %s\n", o.Name(), usrname)
	u, err := o.ReadUser(usrname)

	if err == nil {
		u.Password = ""
	}

	return JSONWriter(writer, u, err)
}

//UpdateUser updates the metadata for existing user from a REST API request
func UpdateUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := strings.ToLower(mux.Vars(request)["user"])
	log.Printf("%s: Update User %s\n", o.Name(), usrname)
	u, err := o.ReadUser(usrname)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}

	modusr := *u
	err = UnmarshalRequest(request, &modusr)
	err = ValidName(modusr.Name, err)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return err
	}

	//We use a special procedure for upgrading the password
	if modusr.Password != u.Password {
		modusr.SetNewPassword(modusr.Password)
	}
	if err = o.UpdateUser(usrname, &modusr); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}
	return JSONWriter(writer, modusr, err)
}

//DeleteUser deletes existing user from a REST API request
func DeleteUser(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := strings.ToLower(mux.Vars(request)["user"])
	log.Printf("%s: Delete User %s\n", o.Name(), usrname)
	err := o.DeleteUser(usrname)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}
	return OK(writer)
}
