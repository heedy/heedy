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

const (
	favicon = `iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAMAAAAp4XiDAAAAM1BMVEVAAABpYjN3c
k18dT+Si0uZlG6em5emoFazr57CvpXFwnbRzcje2Nbi48rm5eHu7Or8/vv8t6tBAAAAAXRSTlMAQObYZ
gAAATtJREFUeAHt1N1ugzAMxXHgYAJNYvz+T7sTPqbSxh253LS/1MTN9JO1m3Z/NF3fip/FWilOU+OSQ
qbmLaxJkDQbklIzYQ0kTUfN/wxziW/iXcL0yCV+v4lk/Vz30uNWV9Fswk0SnreE0iNLCEtWXcawD+Uh5
PL1oQjPJOwtNo+LxSVbHEM25WADX41w4RTqJFsfgiYE44cvChLFGxn3zi1yzEwNZDaDhF9rBDzMVJBNd
iJi88yzQuQgnLKtAwm+CVarEC5GITliLCs2DrG1L4SrXwgA/kGiCcpHdB2gfOC+eSPIJAAuBNHMYg+oM
cExzABJea4QiAw9r15EsA0Dh9J5Xkh/s+6pdlKaf6irlFLSdP1l2e61XCk55EzZcfMok0ecfBJjO2G+i
A5xUYweqavIKBzj1zlNXt3Zf19lqDb7kNICQAAAAABJRU5ErkJggg==`
	faviconMime = "image/png"
)

var (
	//The amount of time to wait between each unsuccessful login attempt
	UnsuccessfulLoginWait = 300 * time.Millisecond
)

//APIHandler is a function that handles some part of the REST API given a specific operator on the database.
type APIHandler func(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error

func authenticator(apifunc APIHandler, db *streamdb.Database) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		//Set up the logger for this connection

		//Since an important use case is behind nginx, the following rule is followed:
		//localhost address is not logged if real-ip header exists (since it is from localhost)
		//if real-ip header exists, faddr=address (forwardedAddress) is logged
		//In essence, if behind nginx, there is no need for the addr=blah

		fields := log.Fields{"addr": request.RemoteAddr, "uri": request.URL.String()}
		if realIP := request.Header.Get("X-Real-IP"); realIP != "" {
			fields["faddr"] = realIP
			if !strings.HasPrefix(request.RemoteAddr, "127.0.0.1") && !strings.HasPrefix(request.RemoteAddr, "::1") {
				delete(fields, "addr")
			}
		}

		logger := log.WithFields(fields)

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

func serveFavicon(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", faviconMime)
	w.Header().Set("Content-Transfer-Encoding", "BASE64")

	w.Write([]byte(favicon))
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	// Special items
	prefix.HandleFunc("/", authenticator(ListUsers, db)).Queries("q", "ls")
	prefix.HandleFunc("/", authenticator(GetThis, db)).Queries("q", "this")
	prefix.HandleFunc("/favicon.ico", serveFavicon)

	prefix.HandleFunc("/", authenticator(RunWebsocket, db)).Headers("Upgrade", "websocket").Methods("GET")

	//User CRUD
	prefix.HandleFunc("/{user}", authenticator(ListDevices, db)).Methods("GET").Queries("q", "ls")
	prefix.HandleFunc("/{user}", authenticator(ReadUser, db)).Methods("GET")
	prefix.HandleFunc("/{user}", authenticator(CreateUser, db)).Methods("POST")
	prefix.HandleFunc("/{user}", authenticator(UpdateUser, db)).Methods("PUT")
	prefix.HandleFunc("/{user}", authenticator(DeleteUser, db)).Methods("DELETE")

	//Device CRUD
	prefix.HandleFunc("/{user}/{device}", authenticator(ListStreams, db)).Methods("GET").Queries("q", "ls")
	prefix.HandleFunc("/{user}/{device}", authenticator(ReadDevice, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}", authenticator(CreateDevice, db)).Methods("POST")
	prefix.HandleFunc("/{user}/{device}", authenticator(UpdateDevice, db)).Methods("PUT")
	prefix.HandleFunc("/{user}/{device}", authenticator(DeleteDevice, db)).Methods("DELETE")

	//Stream CRUD
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(ReadStream, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(CreateStream, db)).Methods("POST")
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(UpdateStream, db)).Methods("PUT")
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(DeleteStream, db)).Methods("DELETE")

	//Stream IO
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(WriteStream, db)).Methods("UPDATE")

	prefix.HandleFunc("/{user}/{device}/{stream}/data", authenticator(GetStreamRangeI, db)).Methods("GET").Queries("i1", "{i1:[0-9]+}")
	prefix.HandleFunc("/{user}/{device}/{stream}/data", authenticator(GetStreamRangeT, db)).Methods("GET").Queries("t1", "{t1:[0-9]*\\.?[0-9]+([eE][-+]?[0-9]+)?}")

	prefix.HandleFunc("/{user}/{device}/{stream}/length", authenticator(GetStreamLength, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}/{stream}/time2index", authenticator(StreamTime2Index, db)).Methods("GET")

	return prefix
}
