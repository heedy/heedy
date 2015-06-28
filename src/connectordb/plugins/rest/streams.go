package rest

import (
	"connectordb/streamdb/operator"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

func getStreamPath(request *http.Request) (username string, devicename string, streamname string, streampath string) {
	username = mux.Vars(request)["user"]
	devicename = mux.Vars(request)["device"]
	streamname = mux.Vars(request)["stream"]
	streampath = username + "/" + devicename + "/" + streamname
	return username, devicename, streamname, streampath
}

//ListStreams lists the streams that the given device has
func ListStreams(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)
	logger = logger.WithField("op", "ListStreams")
	logger.Debugln()
	d, err := o.ReadAllStreams(devpath)
	return JSONWriter(writer, d, logger, err)
}

//CreateStream creates a new stream from a REST API request
func CreateStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, streamname, streampath := getStreamPath(request)
	logger = logger.WithField("op", "CreateStream")
	logger.Infoln()

	err := ValidName(streamname, nil)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		StatsAddFail(err)
		return err
	}

	defer request.Body.Close()

	//Limit the schema to 512KB
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, 512000))
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}

	if err = o.CreateStream(streampath, string(data)); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}

	return ReadStream(o, writer, request, logger)

}

//ReadStream reads a stream from a REST API request
func ReadStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := getStreamPath(request)

	if err := BadQ(o, writer, request, logger); err != nil {
		return err
	}

	logger = logger.WithField("op", "ReadStream")
	logger.Debugln()
	s, err := o.ReadStream(streampath)

	return JSONWriter(writer, s, logger, err)
}

//UpdateStream updates the metadata for existing stream from a REST API request
func UpdateStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := getStreamPath(request)
	logger = logger.WithField("op", "UpdateStream")
	logger.Infoln()

	s, err := o.ReadStream(streampath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	err = UnmarshalRequest(request, s)
	err = ValidName(s.Name, err)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}
	if err = o.UpdateStream(s); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		StatsAddFail(err)
		return err
	}
	return JSONWriter(writer, s, logger, err)
}

//DeleteStream deletes existing stream from a REST API request
func DeleteStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := getStreamPath(request)
	logger = logger.WithField("op", "DeleteStream")
	logger.Infoln()
	err := o.DeleteStream(streampath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		StatsAddFail(err)
		return err
	}
	return OK(writer)
}
