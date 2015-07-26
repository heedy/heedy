package restcore

import (
	"connectordb/streamdb/operator"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"

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
func JSONWriter(writer http.ResponseWriter, data interface{}, logger *log.Entry, err error) (int, string) {
	if err != nil {
		return WriteError(writer, logger, http.StatusNotFound, err, false)
	}

	res, err := json.Marshal(data)
	if err != nil {
		return WriteError(writer, logger, http.StatusInternalServerError, err, true)

	}
	writer.Header().Set("Content-Length", strconv.Itoa(len(res)))
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
	return 0, ""
}

//IntWriter writes an integer
func IntWriter(writer http.ResponseWriter, i int64, logger *log.Entry, err error) (int, string) {
	if err != nil {
		return WriteError(writer, logger, http.StatusForbidden, err, false)
	}

	res := []byte(strconv.FormatInt(i, 10))
	writer.Header().Set("Content-Length", strconv.Itoa(len(res)))
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
	return 0, ""
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
func BadQ(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	if val := request.URL.Query().Get("q"); val != "" {
		return ErrBadQ
	}
	return nil
}

//ErrorResponse is the struct which holds the error message and response code
type ErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"msg"`
	Reference string `json:"ref,omitempty"`
}

//WriteError takes care of gracefully writing errors to the client in a way that allows
//for fairly easy debugging.
func WriteError(writer http.ResponseWriter, logger *log.Entry, errorCode int, err error, iserr bool) (int, string) {
	atomic.AddUint32(&StatsErrors, 1)

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	u, err2 := uuid.NewV4()
	if err2 != nil {
		logger.WithField("ref", "OSHIT").Errorln("Failed to generate error UUID: " + err2.Error())
		logger.WithField("ref", "OSHIT").Warningln("Original Error: " + err.Error())
		writer.WriteHeader(520)
		writer.Write([]byte(`{"code": 520, "msg": "Failed to generate error UUID", "ref": "OSHIT"}`))
		return 1, ""
	}
	uu := u.String()

	response := ErrorResponse{
		Code:      errorCode,
		Message:   err.Error(),
		Reference: uu,
	}
	res, err2 := json.Marshal(response)
	if err2 != nil {
		logger.WithField("ref", uu).Errorln("Failed to marshal error struct: " + err2.Error())
		logger.WithField("ref", uu).Warningln("Original Error: " + err.Error())
		writer.WriteHeader(520)
		writer.Write([]byte(`{"code": 520, "msg": "Failed to write error message","ref":"` + uu + `"}`))
		return 1, ""
	}

	//Now that we have the error message, we log it and send the messages
	l := logger.WithFields(log.Fields{"ref": uu, "code": errorCode})
	if iserr {
		l.Errorln(err.Error())
	} else {
		l.Warningln(err.Error())
	}
	writer.Header().Set("Content-Length", strconv.Itoa(len(res)))
	writer.WriteHeader(errorCode)
	writer.Write(res)
	return 1, ""
}

//GetStreamPath returns the relevant parts of a stream path
func GetStreamPath(request *http.Request) (username string, devicename string, streamname string, streampath string) {
	username = mux.Vars(request)["user"]
	devicename = mux.Vars(request)["device"]
	streamname = mux.Vars(request)["stream"]
	streampath = username + "/" + devicename + "/" + streamname
	return username, devicename, streamname, streampath
}
