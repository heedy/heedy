package crud

import (
	"connectordb/plugins/rest/restcore"
	"connectordb/streamdb"

	"github.com/gorilla/mux"
)

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.HandleFunc("/", restcore.Authenticator(ListUsers, db)).Queries("q", "ls")
	prefix.HandleFunc("/", restcore.Authenticator(GetThis, db)).Queries("q", "this")

	//User CRUD
	prefix.HandleFunc("/{user}", restcore.Authenticator(ListDevices, db)).Methods("GET").Queries("q", "ls")
	prefix.HandleFunc("/{user}", restcore.Authenticator(ReadUser, db)).Methods("GET")
	prefix.HandleFunc("/{user}", restcore.Authenticator(CreateUser, db)).Methods("POST")
	prefix.HandleFunc("/{user}", restcore.Authenticator(UpdateUser, db)).Methods("PUT")
	prefix.HandleFunc("/{user}", restcore.Authenticator(DeleteUser, db)).Methods("DELETE")

	//Device CRUD
	prefix.HandleFunc("/{user}/{device}", restcore.Authenticator(ListStreams, db)).Methods("GET").Queries("q", "ls")
	prefix.HandleFunc("/{user}/{device}", restcore.Authenticator(ReadDevice, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}", restcore.Authenticator(CreateDevice, db)).Methods("POST")
	prefix.HandleFunc("/{user}/{device}", restcore.Authenticator(UpdateDevice, db)).Methods("PUT")
	prefix.HandleFunc("/{user}/{device}", restcore.Authenticator(DeleteDevice, db)).Methods("DELETE")

	//Stream CRUD
	prefix.HandleFunc("/{user}/{device}/{stream}", restcore.Authenticator(ReadStream, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}/{stream}", restcore.Authenticator(CreateStream, db)).Methods("POST")
	prefix.HandleFunc("/{user}/{device}/{stream}", restcore.Authenticator(UpdateStream, db)).Methods("PUT")
	prefix.HandleFunc("/{user}/{device}/{stream}", restcore.Authenticator(DeleteStream, db)).Methods("DELETE")

	//Stream IO
	prefix.HandleFunc("/{user}/{device}/{stream}", restcore.Authenticator(WriteStream, db)).Methods("UPDATE") //Restamp off
	prefix.HandleFunc("/{user}/{device}/{stream}", restcore.Authenticator(WriteStream, db)).Methods("PATCH")  //Restamp on

	prefix.HandleFunc("/{user}/{device}/{stream}/data", restcore.Authenticator(GetStreamRange, db)).Methods("GET")

	prefix.HandleFunc("/{user}/{device}/{stream}/length", restcore.Authenticator(GetStreamLength, db)).Methods("GET")
	prefix.HandleFunc("/{user}/{device}/{stream}/time2index", restcore.Authenticator(StreamTime2Index, db)).Methods("GET")

	return prefix
}
