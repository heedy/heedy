package rest

import (
	"connectordb/streamdb/operator"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//Mb is nubmer of bytes in a megabyte
const Mb = 1024 * 1024

//ErrInvalidName is thrown when the name is bad
var (
	ErrInvalidName = errors.New("The given name did not pass sanitation.")
	ErrBadQ        = errors.New("Unrecognized query command.")
)

//OK is a simplifying function that returns success
func OK(writer http.ResponseWriter) error {
	writer.Header().Set("Content-Length", "2")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("ok"))
	return nil
}

//JSONWriter writes the given data as http
func JSONWriter(writer http.ResponseWriter, data interface{}, logger *log.Entry, err error) error {
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		logger.Warningln(err)
		return err
	}

	res, err := json.Marshal(data)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Errorln(err)
		return err
	}
	writer.Header().Set("Content-Length", strconv.Itoa(len(res)))
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
	return nil
}

//UnmarshalRequest unmarshals the input data to the given interface
func UnmarshalRequest(request *http.Request, unmarshalTo interface{}) error {
	defer request.Body.Close()

	//Limit requests to 10MB
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, 10*Mb))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, unmarshalTo)
}

//ValidName sanitizes names so that only valid ones are added
func ValidName(n string, err error) error {
	if err != nil {
		return err
	}

	if strings.Contains(n, "/") ||
		strings.Contains(n, "\\") ||
		strings.Contains(n, " ") ||
		strings.Contains(n, "?") ||
		len(n) == 0 {
		return ErrInvalidName
	}

	return nil
}

//BadQ checks if there is a q= part to the given query, and gives an error if there is
func BadQ(o operator.Operator, writer http.ResponseWriter, request *http.Request, arg string) error {
	if val := request.URL.Query().Get("q"); val != "" {
		writer.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr,
			"op": "Q", "arg": arg}).Warningln("Bad Q: ", val)
		return ErrBadQ
	}
	return nil
}
