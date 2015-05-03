package rest

import (
	"net/http"
	"streamdb"
)

//GetThis is a command to return the "username/devicename" of the currently authenticated thing
func GetThis(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usr, err := o.User()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return err
	}
	dev, err := o.Device()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return err
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(usr.Name + "/" + dev.Name))
	return nil
}

//CreateDevice creates a new user from a REST API request
func CreateDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}

//GetDevice gets an existing user from a REST API request
func GetDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}

//UpdateDevice updates the metadata for existing user from a REST API request
func UpdateDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}

//DeleteDevice deletes existing user from a REST API request
func DeleteDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}
