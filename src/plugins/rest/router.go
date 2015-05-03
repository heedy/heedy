package rest

import (
	"streamdb"

	"errors"
	"net/http"

	"github.com/gorilla/mux"
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
			return
		}

		//Now, authentication is either by API key or by username/password combo
		var o streamdb.Operator
		var err error

		if len(authUser) != 0 {
			//Authenticate by username/password
			o, err = db.UserOperator(authUser, authPass)

			if err != nil {
				writer.Header().Set("WWW-Authenticate", "Basic")
				writer.WriteHeader(http.StatusUnauthorized)
				writer.Write([]byte(err.Error()))
				return
			}
		} else {
			//Authenticate by API key
			o, err = db.DeviceOperator(authPass)

			if err != nil {
				writer.Header().Set("WWW-Authenticate", "Basic")
				writer.WriteHeader(http.StatusUnauthorized)
				writer.Write([]byte(err.Error()))
				return
			}
		}

		//If we got here, o is valid.
		err = apifunc(o, writer, request)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
	})
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	//Data-handlers (CRUD)
	d := prefix.PathPrefix("/d").Subrouter()

	//User CRUD
	d.HandleFunc("/{user}", authenticator(GetUser, db)).Methods("GET")
	d.HandleFunc("/{user}", authenticator(CreateUser, db)).Methods("POST")
	d.HandleFunc("/{user}", authenticator(UpdateUser, db)).Methods("PUT")
	d.HandleFunc("/{user}", authenticator(DeleteUser, db)).Methods("DELETE")

	//Device CRUD
	d.HandleFunc("/{user}/{device}", authenticator(GetDevice, db)).Methods("GET")
	d.HandleFunc("/{user}/{device}", authenticator(CreateDevice, db)).Methods("POST")
	d.HandleFunc("/{user}/{device}", authenticator(UpdateDevice, db)).Methods("PUT")
	d.HandleFunc("/{user}/{device}", authenticator(DeleteDevice, db)).Methods("DELETE")

	//Stream CRUD
	d.HandleFunc("/{user}/{device}/{stream}", authenticator(GetStream, db)).Methods("GET")
	d.HandleFunc("/{user}/{device}/{stream}", authenticator(CreateStream, db)).Methods("POST")
	d.HandleFunc("/{user}/{device}/{stream}", authenticator(UpdateStream, db)).Methods("PUT")
	d.HandleFunc("/{user}/{device}/{stream}", authenticator(DeleteStream, db)).Methods("DELETE")

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
