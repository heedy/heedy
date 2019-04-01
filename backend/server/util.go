package server

import (
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"encoding/json"

	"net/http"

	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

// apiHeaders writes headers that need to be present in all API requests
func apiHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // All API requests return json
}

//UnmarshalRequest unmarshals the input data to the given interface
func UnmarshalRequest(request *http.Request, unmarshalTo interface{}) error {
	defer request.Body.Close()

	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, *assets.Config().RequestBodyByteLimit))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, unmarshalTo)
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ID               string `json:"id,omitempty"`
}

// WriteJSONError writes an error message as json. It is assumed that the resulting
// status code is not StatusOK, but rather 4xx
func WriteJSONError(w http.ResponseWriter, r *http.Request, status int, err error) {
	c := CTX(r)

	es := ErrorResponse{
		Error:            "access_denied",
		ErrorDescription: err.Error(),
	}

	// We can have error types encoded in the error, split with a :
	errs := strings.SplitN(err.Error(), ":", 2)
	if len(errs) > 1 && !strings.Contains(errs[0], " ") {
		es.Error = errs[0]
		es.ErrorDescription = strings.TrimSpace(errs[1])
	}

	if c != nil {
		es.ID = c.ID
	}
	jes, err := json.Marshal(&es)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server_error", "error_description": "Failed to create error message"}`))
		if c != nil {
			c.Log.Errorf("Failed to write error message: %s", err)
		} else {
			logrus.Errorf("Failed to write error message: %s", err)
		}
	}

	if c != nil {
		c.Log.Warn(err)
	} else {
		logrus.Warn(err)
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jes)))
	w.WriteHeader(status)
	w.Write(jes)
}

// WriteJSON writes response as JSON, or writes the error if such is given
func WriteJSON(w http.ResponseWriter, r *http.Request, data interface{}, err error) {
	if err != nil {
		// By default, an error returns 400
		WriteJSONError(w, r, 400, err)
		return
	}
	jdata, err := json.Marshal(data)
	if err != nil {
		WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jdata)))
	w.WriteHeader(http.StatusOK)
	w.Write(jdata)
}

// WriteResult writes "ok" if the command succeeded, and outputs an error if it didn't
func WriteResult(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		// By default, an error returns 400
		WriteJSONError(w, r, 400, err)
		return
	}
	// success :)
	w.Header().Set("Content-Length", "4")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`"ok"`))

}
