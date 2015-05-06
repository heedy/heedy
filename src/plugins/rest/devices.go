package rest

import (
	"net/http"
	"streamdb"
)

//GetThis is a command to return the "username/devicename" of the currently authenticated thing
func GetThis(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(o.Name()))
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
