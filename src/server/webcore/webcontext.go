package webcore

import (
	"connectordb/operator"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Provides utilities and information about the current web request.
type WebContext struct {
	Operator operator.Operator
	Writer   http.ResponseWriter
	Request  *http.Request
	Logger   *log.Entry
}

// Returns the username submitted by the request
func (wc *WebContext) GetUsername() string {
	return mux.Vars(wc.Request)["user"]
}

// Returns the device name submitted by the request
func (wc *WebContext) GetDeviceName() string {
	return mux.Vars(wc.Request)["device"]
}

// Returns the stream name submitted by the request
func (wc *WebContext) GetStreamName() string {
	return mux.Vars(wc.Request)["stream"]
}

func (wc *WebContext) GetDevicePath() string {
	return wc.GetUsername() + "/" + wc.GetDeviceName()
}

func (wc *WebContext) GetStreamPath() string {
	return wc.GetDevicePath() + "/" + wc.GetStreamName()
}
