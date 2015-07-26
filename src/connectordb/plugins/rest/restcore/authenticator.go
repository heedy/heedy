package restcore

import (
	"connectordb/streamdb"
	"connectordb/streamdb/operator"
	"errors"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	//UnsuccessfulLoginWait is the amount of time to wait between each unsuccessful login attempt
	UnsuccessfulLoginWait = 300 * time.Millisecond
)

//GetRequestLogger returns a logrus log entry which has fields prepopulated for the request
func GetRequestLogger(request *http.Request, opname string) *log.Entry {
	//Since an important use case is behind nginx, the following rule is followed:
	//localhost address is not logged if real-ip header exists (since it is from localhost)
	//if real-ip header exists, faddr=address (forwardedAddress) is logged
	//In essence, if behind nginx, there is no need for the addr=blah

	fields := log.Fields{"addr": request.RemoteAddr, "uri": request.URL.Path, "op": opname}
	if realIP := request.Header.Get("X-Real-IP"); realIP != "" {
		fields["faddr"] = realIP
		if strings.HasPrefix(request.RemoteAddr, "127.0.0.1") || strings.HasPrefix(request.RemoteAddr, "::1") {
			delete(fields, "addr")
		}
	}

	return log.WithFields(fields)
}

//WriteAccessControlHeaders writes the access control headers for the site
func WriteAccessControlHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, UPDATE, PATCH")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

//APIHandler is a function that handles some part of the REST API given a specific operator on the database.
type APIHandler func(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error

//Authenticator is a wrapper function that sets up authentication and database for each request
func Authenticator(apifunc APIHandler, db *streamdb.Database) http.HandlerFunc {
	funcname := runtime.FuncForPC(reflect.ValueOf(apifunc).Pointer()).Name()

	//funcname is a full path - to simplify logs, we split it into just the function name, assuming that function names are strictly unique
	funcname = strings.Split(funcname, ".")[1]

	//Sets up the query timer for this api call if it doesn't exist yet
	qtimer, ok := QueryTimers[funcname]
	if !ok {
		qtimer = &QueryTimer{}
		QueryTimers[funcname] = qtimer
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		tstart := time.Now()

		//Increment the queries handled
		atomic.AddUint32(&StatsQueries, 1)

		//Set up the logger for this connection
		logger := GetRequestLogger(request, funcname)

		WriteAccessControlHeaders(writer)

		//Check authentication
		authUser, authPass, ok := request.BasicAuth()

		if !ok {
			//If there is no basic auth header, check for apikey parameter in the query itself
			authPass = request.URL.Query().Get("apikey")

			//If there was no apikey, fail asking for basic auth - otherwise, continue login with the api key
			if len(authPass) == 0 {
				writer.Header().Set("WWW-Authenticate", "Basic")
				WriteError(writer, logger, http.StatusUnauthorized, errors.New("Login attempted without authentication"), false)
				atomic.AddUint32(&StatsAuthFails, 1)
				return
			}
		}

		//Handle a panic without crashing the whole rest interface
		defer func() {
			if r := recover(); r != nil {
				atomic.AddUint32(&StatsPanics, 1)
				logger.WithField("dev", authUser).Errorln("PANIC: " + r.(error).Error())
			}
		}()

		o, err := db.LoginOperator(authUser, authPass)

		if err != nil {
			//So there was an unsuccessful attempt at login, huh?
			time.Sleep(UnsuccessfulLoginWait)

			WriteError(writer, logger, http.StatusUnauthorized, err, false)
			atomic.AddUint32(&StatsAuthFails, 1)

			return
		}
		l := logger.WithField("dev", o.Name())

		//If we got here, o is a valid operator
		apifunc(o, writer, request, l)

		//Write the time that this query took
		tdiff := time.Since(tstart)
		qtimer.Add(tdiff)
		l.Debugln(tdiff)
	})
}
