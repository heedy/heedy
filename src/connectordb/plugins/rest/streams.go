package rest

import (
	"connectordb/streamdb"
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
func ListStreams(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, devpath := getDevicePath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "ListStreams", "arg": devpath})
	logger.Debugln()
	d, err := o.ReadAllStreams(devpath)
	return JSONWriter(writer, d, logger, err)
}

//CreateStream creates a new stream from a REST API request
func CreateStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, streamname, streampath := getStreamPath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "CreateStream", "arg": streampath})
	logger.Infoln()

	err := ValidName(streamname, nil)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
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

	return ReadStream(o, writer, request)

}

//ReadStream reads a stream from a REST API request
func ReadStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, _, streampath := getStreamPath(request)

	if err := BadQ(o, writer, request, streampath); err != nil {
		return err
	}

	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "ReadStream", "arg": streampath})
	logger.Debugln()
	s, err := o.ReadStream(streampath)

	return JSONWriter(writer, s, logger, err)
}

//UpdateStream updates the metadata for existing stream from a REST API request
func UpdateStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, _, streampath := getStreamPath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "UpdateStream", "arg": streampath})
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
		return err
	}
	return JSONWriter(writer, s, logger, err)
}

//DeleteStream deletes existing stream from a REST API request
func DeleteStream(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, _, streampath := getStreamPath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "DeleteStream", "arg": streampath})
	logger.Infoln()
	err := o.DeleteStream(streampath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	return OK(writer)
}
