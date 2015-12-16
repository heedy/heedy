/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package webcore

import (
	"connectordb/operator"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"sync/atomic"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
)

var (
	//The name of the site.
	SiteName string
	//AllowCrossOrigin: Whether or not cross origin requests are permitted
	AllowCrossOrigin = false

	//IsActive - no need for sync, really. It specifies if the server should accept connections.
	IsActive = true

	//ShutdownChannel is a shared channel which is used when a shutdown is signalled.
	//Each goroutine that uses the ShutdownChannel is to IMMEDIATELY refire the channel before doing anything else,
	//so that the signal continues throughout the system
	ShutdownChannel = make(chan bool, 1)
)

//APIHandler is a function that handles some part of the REST API given a specific operator on the database.
type APIHandler func(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string)

//WriteAccessControlHeaders writes the access control headers for the site
func WriteAccessControlHeaders(writer http.ResponseWriter, request *http.Request) {
	if AllowCrossOrigin {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		//writer.Header().Set("Access-Control-Allow-Origin", SiteName)
	}
	//Only set allow credentials if the origin is the site name
	originheader := request.Header.Get("Origin")
	if originheader == "" || originheader == SiteName {
		//Only permit cookies if we are coming from our own origin
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

//SetEnabled allows to enable and disable acceptance of connections in a simple way
func SetEnabled(v bool) {
	IsActive = v
	if v {
		log.Warn("Web server enabled")
	} else {
		log.Warn("Web server disabled (503)")
	}
}

//Shutdown shutd down the server
func Shutdown() {
	//Set to inactive so that new connections are not accepted during shutdown
	//no need to log the fact that rest is inactive, since this only happens on shutdown
	IsActive = false
	//Fire the shutdown channel
	ShutdownChannel <- true
}

//GetStreamPath returns the relevant parts of a stream path
func GetStreamPath(request *http.Request) (username string, devicename string, streamname string, streampath string) {
	username = mux.Vars(request)["user"]
	devicename = mux.Vars(request)["device"]
	streamname = mux.Vars(request)["stream"]
	streampath = username + "/" + devicename + "/" + streamname
	return username, devicename, streamname, streampath
}

//GetFuncName returns the name of the function that is going to handle a request
func GetFuncName(apifunc APIHandler) string {
	funcname := runtime.FuncForPC(reflect.ValueOf(apifunc).Pointer()).Name()

	//funcname is a full path - to simplify logs, we split it into just the function name, assuming that function names are strictly unique
	return strings.Split(funcname, ".")[1]
}

//HandlePanic is called in defer statements to handle a panic within a request.
//It is assumed that the connection is active
func HandlePanic(logger *log.Entry) {
	if r := recover(); r != nil {
		atomic.AddInt32(&StatsActive, -1)
		atomic.AddUint32(&StatsPanics, 1)
		logger.Errorf("PANIC: %s\n\n%s\n\n", r.(error).Error(), debug.Stack())
	}
}
