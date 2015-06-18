package rest

import (
	"connectordb/streamdb"
	"connectordb/streamdb/operator"
	"time"

	log "github.com/Sirupsen/logrus"

	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var (
	//UnsuccessfulLoginWait is the amount of time to wait between each unsuccessful login attempt
	UnsuccessfulLoginWait = 300 * time.Millisecond
)

func getLogger(request *http.Request) *log.Entry {
	//Since an important use case is behind nginx, the following rule is followed:
	//localhost address is not logged if real-ip header exists (since it is from localhost)
	//if real-ip header exists, faddr=address (forwardedAddress) is logged
	//In essence, if behind nginx, there is no need for the addr=blah

	fields := log.Fields{"addr": request.RemoteAddr, "uri": request.URL.String()}
	if realIP := request.Header.Get("X-Real-IP"); realIP != "" {
		fields["faddr"] = realIP
		if strings.HasPrefix(request.RemoteAddr, "127.0.0.1") || strings.HasPrefix(request.RemoteAddr, "::1") {
			delete(fields, "addr")
		}
	}

	return log.WithFields(fields)
}

//Writes the access control headers for the site
func writeAccessControlHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, UPDATE")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

//APIHandler is a function that handles some part of the REST API given a specific operator on the database.
type APIHandler func(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error

func authenticator(apifunc APIHandler, db *streamdb.Database) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		//Set up the logger for this connection
		logger := getLogger(request)

		writeAccessControlHeaders(writer)

		//Check authentication
		authUser, authPass, ok := request.BasicAuth()

		//If there is no basic auth header, return unauthorized
		if !ok {
			writer.Header().Set("WWW-Authenticate", "Basic")
			writer.WriteHeader(http.StatusUnauthorized)
			logger.WithField("op", "AUTH").Warningln("Login attempt w/o auth")
			return
		}

		//Handle a panic without crashing the whole rest interface
		defer func() {
			if r := recover(); r != nil {
				logger.WithFields(log.Fields{"dev": authUser, "op": "PANIC"}).Errorln(r)
			}
		}()

		o, err := db.LoginOperator(authUser, authPass)

		if err != nil {
			logger.WithFields(log.Fields{"dev": authUser, "op": "AUTH"}).Warningln(err.Error())

			//So there was an unsuccessful attempt at login, huh?
			time.Sleep(UnsuccessfulLoginWait)

			writer.Header().Set("WWW-Authenticate", "Basic")
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte(err.Error()))

			return
		}

		//If we got here, o is a valid operator
		err = apifunc(o, writer, request, logger.WithField("dev", o.Name()))
		if err != nil {
			writer.Write([]byte(err.Error()))
		}
	})
}

//When a path is not found, return a 404 with path not recognized message
func notfoundHandler(writer http.ResponseWriter, request *http.Request) {
	getLogger(request).WithField("method", request.Method).Debug("404")
	writer.WriteHeader(http.StatusNotFound)
	writer.Write([]byte("This path is not recognized"))

}

//on OPTIONS to allow cross-site XMLHTTPRequest, allow access control origin
func optionsHandler(writer http.ResponseWriter, request *http.Request) {
	getLogger(request).WithField("method", request.Method).Debug()
	writeAccessControlHeaders(writer)
	writer.WriteHeader(http.StatusOK)
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.NotFoundHandler = http.HandlerFunc(notfoundHandler)

	prefix.Methods("OPTIONS").Handler(http.HandlerFunc(optionsHandler))

	// Special items
	prefix.HandleFunc("/", authenticator(RunWebsocket, db)).Headers("Upgrade", "websocket").Methods("GET")

	//The 'd' prefix corresponds to data
	d := prefix.PathPrefix("/d").Subrouter()

	d.HandleFunc("/", authenticator(ListUsers, db)).Queries("q", "ls")
	d.HandleFunc("/", authenticator(GetThis, db)).Queries("q", "this")

	//User CRUD
	userPath := "/{user}"
	d.HandleFunc(userPath, authenticator(ListDevices, db)).Methods("GET").Queries("q", "ls")
	d.HandleFunc(userPath, authenticator(ReadUser, db)).Methods("GET")
	d.HandleFunc(userPath, authenticator(CreateUser, db)).Methods("POST")
	d.HandleFunc(userPath, authenticator(UpdateUser, db)).Methods("PUT")
	d.HandleFunc(userPath, authenticator(DeleteUser, db)).Methods("DELETE")

	//Device CRUD
	devicePath := userPath + "/{device}"
	d.HandleFunc(devicePath, authenticator(ListStreams, db)).Methods("GET").Queries("q", "ls")
	d.HandleFunc(devicePath, authenticator(ReadDevice, db)).Methods("GET")
	d.HandleFunc(devicePath, authenticator(CreateDevice, db)).Methods("POST")
	d.HandleFunc(devicePath, authenticator(UpdateDevice, db)).Methods("PUT")
	d.HandleFunc(devicePath, authenticator(DeleteDevice, db)).Methods("DELETE")

	//Stream CRUD
	streamPath := devicePath + "/{stream}"
	d.HandleFunc(streamPath, authenticator(ReadStream, db)).Methods("GET")
	d.HandleFunc(streamPath, authenticator(CreateStream, db)).Methods("POST")
	d.HandleFunc(streamPath, authenticator(UpdateStream, db)).Methods("PUT")
	d.HandleFunc(streamPath, authenticator(DeleteStream, db)).Methods("DELETE")

	//Stream IO
	d.HandleFunc(streamPath, authenticator(WriteStream, db)).Methods("UPDATE")

	d.HandleFunc(streamPath+"/data", authenticator(GetStreamRangeI, db)).Methods("GET").Queries("i1", "{i1}")
	d.HandleFunc(streamPath+"/data", authenticator(GetStreamRangeT, db)).Methods("GET").Queries("t1", "{t1}")

	d.HandleFunc(streamPath+"/length", authenticator(GetStreamLength, db)).Methods("GET")
	d.HandleFunc(streamPath+"/time2index", authenticator(StreamTime2Index, db)).Methods("GET")

	return prefix
}
