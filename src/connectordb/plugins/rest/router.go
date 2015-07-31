package rest

import (
	"connectordb/streamdb"
	"connectordb/streamdb/operator"
	"errors"

	"net/http"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/gorilla/mux"

	"connectordb/plugins/rest/crud"
	"connectordb/plugins/rest/dataset"
	"connectordb/plugins/rest/feed"
	"connectordb/plugins/rest/restcore"

	log "github.com/Sirupsen/logrus"
)

var (
	//PreferredFileLimit sets the preferred maximum number of open files
	PreferredFileLimit = uint64(10000)
)

//SetFileLimit attempts to set the open file limits
func SetFileLimit() {
	var noFile syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &noFile)
	if err != nil {
		log.Warn("Could not read file limit:", err)
		return
	}
	if noFile.Cur < PreferredFileLimit {
		change := uint64(0)
		if noFile.Max < PreferredFileLimit {
			change = noFile.Max
			log.Warnf("User hard file limit (%d) is less than preferred %d", noFile.Max, PreferredFileLimit)
		} else {
			change = PreferredFileLimit
		}
		log.Warnf("Setting user file limit from %d to %d", noFile.Cur, change)
		noFile.Cur = change
		if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &noFile); err != nil {
			log.Error("Failed to set file limit: ", err)
		}
	}
}

//NotFoundHandler when a path is not found, return a 404 with path not recognized message
func NotFoundHandler(writer http.ResponseWriter, request *http.Request) {
	atomic.AddUint32(&restcore.StatsQueries, 1)
	logger := restcore.GetRequestLogger(request, "404")
	restcore.WriteError(writer, logger, http.StatusNotFound, errors.New("This path is not recognized"), false)
}

//OptionsHandler on OPTIONS to allow cross-site XMLHTTPRequest, allow access control origin
func OptionsHandler(writer http.ResponseWriter, request *http.Request) {
	atomic.AddUint32(&restcore.StatsQueries, 1)
	restcore.GetRequestLogger(request, "OPTIONS").Debug()
	restcore.WriteAccessControlHeaders(writer)
	writer.WriteHeader(http.StatusOK)
}

//GetThis is a command to return the "username/devicename" of the currently authenticated thing
func GetThis(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	res := []byte(o.Name())
	writer.Header().Set("Content-Length", strconv.Itoa(len(res)))
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
	return restcore.DEBUG, ""
}

//CountAllUsers gets all of the users in the entire database
func CountAllUsers(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	l, err := o.CountUsers()
	return restcore.UintWriter(writer, l, logger, err)
}

//CountAllDevices gets all of the devices in the entire database
func CountAllDevices(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	l, err := o.CountDevices()
	return restcore.UintWriter(writer, l, logger, err)
}

//CountAllStreams gets all of the streams in the entire database
func CountAllStreams(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	l, err := o.CountStreams()
	return restcore.UintWriter(writer, l, logger, err)
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	SetFileLimit()

	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	prefix.Methods("OPTIONS").Handler(http.HandlerFunc(OptionsHandler))

	// The websocket is run straight from here
	prefix.HandleFunc("/", restcore.Authenticator(RunWebsocket, db)).Headers("Upgrade", "websocket").Methods("GET")

	prefix.HandleFunc("/", restcore.Authenticator(GetThis, db)).Queries("q", "this").Methods("GET")
	prefix.HandleFunc("/", restcore.Authenticator(CountAllUsers, db)).Queries("q", "countusers").Methods("GET")
	prefix.HandleFunc("/", restcore.Authenticator(CountAllDevices, db)).Queries("q", "countdevices").Methods("GET")
	prefix.HandleFunc("/", restcore.Authenticator(CountAllStreams, db)).Queries("q", "countstreams").Methods("GET")

	crud.Router(db, prefix.PathPrefix("/crud").Subrouter())
	dataset.Router(db, prefix.PathPrefix("/dataset").Subrouter())
	feed.Router(db, prefix.PathPrefix("/feed").Subrouter())

	go restcore.RunStats()
	go restcore.RunQueryTimers()

	return prefix
}
