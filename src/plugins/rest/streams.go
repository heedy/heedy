package rest

import (
	"net/http"
	"streamdb"
)

//CreateStream creates a new user from a REST API request
func CreateStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}

//GetStream gets an existing user from a REST API request
func GetStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}

//UpdateStream updates the metadata for existing user from a REST API request
func UpdateStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}

//DeleteStream deletes existing user from a REST API request
func DeleteStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusNotImplemented)
	return ErrUnderConstruction
}
