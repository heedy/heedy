package rest

import (
	"streamdb"
	"time"

	log "github.com/Sirupsen/logrus"

	"errors"
	"net/http"

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
	//ErrUnderConstruction is returned when an API call is valid, but currently unimplemented
	ErrUnderConstruction = errors.New("This part of the API is under construction.")
)

//APIHandler is a function that handles some part of the REST API given a specific operator on the database.
type APIHandler func(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error

func authenticator(apifunc APIHandler, db *streamdb.Database) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authUser, authPass, ok := request.BasicAuth()

		//If there is no basic auth header, return unauthorized
		if !ok {
			writer.Header().Set("WWW-Authenticate", "Basic")
			writer.WriteHeader(http.StatusUnauthorized)
			log.WithField("op", "AUTH").Warningln("Login attempt w/o auth")
			return
		}

		o, err := db.LoginOperator(authUser, authPass)

		if err != nil {
			log.WithFields(log.Fields{"dev": authUser, "addr": request.RemoteAddr, "op": "AUTH"}).Warningln(err.Error())

			//So there was an unsuccessful attempt at login, huh?
			time.Sleep(300 * time.Millisecond)

			writer.Header().Set("WWW-Authenticate", "Basic")
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte(err.Error()))

			return
		}

		//If we got here, o is a valid operator
		err = apifunc(o, writer, request)
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
	prefix.HandleFunc("/", authenticator(ListUsers, db)).Queries("special", "ls")
	prefix.HandleFunc("/", authenticator(GetThis, db)).Queries("special", "this")
	prefix.HandleFunc("/favicon.ico", serveFavicon)

	//User CRUD
	prefix.HandleFunc("/{user}", authenticator(ListDevices, db)).Methods("GET").Queries("special", "ls")
	prefix.HandleFunc("/{user}", authenticator(ReadUser, db)).Methods("GET")
	prefix.HandleFunc("/{user}", authenticator(CreateUser, db)).Methods("POST")
	prefix.HandleFunc("/{user}", authenticator(UpdateUser, db)).Methods("PUT")
	prefix.HandleFunc("/{user}", authenticator(DeleteUser, db)).Methods("DELETE")

	//Device CRUD
	prefix.HandleFunc("/{user}/{device}", authenticator(ListStreams, db)).Methods("GET").Queries("special", "ls")
	prefix.HandleFunc("/{user}/{device}", authenticator(ReadDevice, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}", authenticator(CreateDevice, db)).Methods("POST")
	prefix.HandleFunc("/{user}/{device}", authenticator(UpdateDevice, db)).Methods("PUT")
	prefix.HandleFunc("/{user}/{device}", authenticator(DeleteDevice, db)).Methods("DELETE")

	//Stream CRUD
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(ReadStream, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(CreateStream, db)).Methods("POST")
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(UpdateStream, db)).Methods("PUT")
	prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(DeleteStream, db)).Methods("DELETE")
	//prefix.HandleFunc("/{user}/{device}/{stream}", authenticator(WriteStream, db)).Methods("UPDATE")

	//Getting details of the stream
	//prefix.HandleFunc("/{user}/{device}/{stream}",<>).Methods("GET")

	//Get an index range - start at index i1, and return i2 datapoints
	//prefix.HandleFunc("/{user}/{device}/{stream}/i/{i1:[0-9]+}/{i2:[0-9]+}",<>).Methods("GET")
	//Get a time range - start at time t1, and end at time t2
	//prefix.HandleFunc("/{user}/{device}/{stream}/t/{t1:[0-9]+}/{t2:[0-9]+}",<>).Methods("GET")
	//Get a time range - start at time t1, and return i2 datapoints
	//prefix.HandleFunc("/{user}/{device}/{stream}/t/{t1:[0-9]+}/{t2:[0-9]+}",<>).Methods("GET")

	//Connect to the device websocket
	//prefix.HandleFunc("/{user}/{device}.ws",<>).Methods("GET")

	//Function Handlers
	//f := prefix.PathPrefix("/f").Subrouter()
	//f.HandleFunc("/this", authenticator(GetThis, db)).Methods("GET")
	//f.HandleFunc("/ls", authenticator(ListUsers, db)).Methods("GET")

	//Future handlers: m (models and machine learning)

	return prefix
}
