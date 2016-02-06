/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package crud

import (
	"connectordb/authoperator"
	"io"
	"io/ioutil"
	"net/http"
	"server/restapi/restcore"
	"server/webcore"

	log "github.com/Sirupsen/logrus"
)

//ListStreams lists the streams that the given device has
func ListStreams(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, devpath := getDevicePath(request)
	d, err := o.ReadDeviceStreamsToMap(devpath)
	return restcore.JSONWriter(writer, d, logger, err)
}

//CreateStream creates a new stream from a REST API request
func CreateStream(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, streamname, streampath := restcore.GetStreamPath(request)

	err := restcore.ValidName(streamname, nil)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	defer request.Body.Close()

	//Limit the schema to 512KB
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, 512000))
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	if err = o.CreateStream(streampath, string(data)); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}

	return ReadStream(o, writer, request, logger)

}

//ReadStream reads a stream from a REST API request
func ReadStream(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	s, err := o.ReadStream(streampath)

	return restcore.JSONWriter(writer, s, logger, err)
}

//UpdateStream updates the metadata for existing stream from a REST API request
func UpdateStream(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)

	var supdate map[string]interface{}
	err := restcore.UnmarshalRequest(request, &supdate)

	if err = o.UpdateStream(streampath, supdate); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	s, err := o.ReadStream(streampath)
	return restcore.JSONWriter(writer, s, logger, err)
}

//DeleteStream deletes existing stream from a REST API request
func DeleteStream(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)

	err := o.DeleteStream(streampath)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	restcore.OK(writer)
	return webcore.DEBUG, ""
}
