/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package crud

import (
	"connectordb/authoperator"
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

	sm, err := o.StreamMaker()
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	err = restcore.UnmarshalRequest(request, sm)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	sm.Name = streamname
	if err = o.CreateStream(streampath, sm); err != nil {
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
