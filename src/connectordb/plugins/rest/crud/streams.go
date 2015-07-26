package crud

import (
	"connectordb/plugins/rest/restcore"
	"connectordb/streamdb/operator"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

//ListStreams lists the streams that the given device has
func ListStreams(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)
	d, err := o.ReadAllStreams(devpath)
	return restcore.JSONWriter(writer, d, logger, err)
}

//CreateStream creates a new stream from a REST API request
func CreateStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, streamname, streampath := restcore.GetStreamPath(request)

	err := restcore.ValidName(streamname, nil)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	defer request.Body.Close()

	//Limit the schema to 512KB
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, 512000))
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	if err = o.CreateStream(streampath, string(data)); err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}

	return ReadStream(o, writer, request, logger)

}

//ReadStream reads a stream from a REST API request
func ReadStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := restcore.GetStreamPath(request)

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}

	s, err := o.ReadStream(streampath)

	return restcore.JSONWriter(writer, s, logger, err)
}

//UpdateStream updates the metadata for existing stream from a REST API request
func UpdateStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := restcore.GetStreamPath(request)

	s, err := o.ReadStream(streampath)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}
	err = restcore.UnmarshalRequest(request, s)
	err = restcore.ValidName(s.Name, err)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return err
	}
	if err = o.UpdateStream(s); err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}
	return restcore.JSONWriter(writer, s, logger, err)
}

//DeleteStream deletes existing stream from a REST API request
func DeleteStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := restcore.GetStreamPath(request)

	//Deleting stream is info-worthy
	logger.Infoln()

	err := o.DeleteStream(streampath)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}
	return restcore.OK(writer)
}
