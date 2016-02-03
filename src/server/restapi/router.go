/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package restapi

import (
	"connectordb"
	"connectordb/authoperator"
	"util"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"server/restapi/crud"
	"server/restapi/feed"
	"server/restapi/meta"
	"server/restapi/query"
	"server/restapi/restcore"
	"server/webcore"

	log "github.com/Sirupsen/logrus"
)

//GetThis is a command to return the "username/devicename" of the currently authenticated thing
func GetThis(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	res := []byte(o.Name())
	writer.Header().Set("Content-Length", strconv.Itoa(len(res)))
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
	return webcore.DEBUG, ""
}

//CountAllUsers gets all of the users in the entire database
func CountAllUsers(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	l, err := o.CountUsers()
	return restcore.UintWriter(writer, uint64(l), logger, err)
}

//CountAllDevices gets all of the devices in the entire database
func CountAllDevices(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	l, err := o.CountDevices()
	return restcore.UintWriter(writer, uint64(l), logger, err)
}

//CountAllStreams gets all of the streams in the entire database
func CountAllStreams(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	l, err := o.CountStreams()
	return restcore.UintWriter(writer, uint64(l), logger, err)
}

//Login handles logging in and out of the web interface. In particular, it handles the auth cookies, and
//the web interface uses them for the rest
func Login(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	err := webcore.CreateSessionCookie(o, writer, request)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
	}
	restcore.OK(writer)

	return webcore.DEBUG, ""

}

//Logout hondles logging out of the web interface. It deletes the auth cookie
func Logout(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	webcore.CreateSessionCookie(nil, writer, request) //nil operator deletes the cookie
	restcore.OK(writer)
	return webcore.DEBUG, ""
}

//Allows to fit the Closer interface
type restcloser struct{}

//CLose shuts down the rest server, and makes sure all websockets have exited
func (r restcloser) Close() {
	webcore.Shutdown() //This is a sort of roundabout way of shutting everything down.
	//might want to refactor the above down a directory level at some point
	websocketWaitGroup.Wait()
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *connectordb.Database, prefix *mux.Router) (*mux.Router, error) {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.HandleFunc("/", restcore.Authenticator(GetThis, db)).Queries("q", "this").Methods("GET")
	prefix.HandleFunc("/", restcore.Authenticator(CountAllUsers, db)).Queries("q", "countusers").Methods("GET")
	prefix.HandleFunc("/", restcore.Authenticator(CountAllDevices, db)).Queries("q", "countdevices").Methods("GET")
	prefix.HandleFunc("/", restcore.Authenticator(CountAllStreams, db)).Queries("q", "countstreams").Methods("GET")

	// The websocket is run straight from here
	prefix.HandleFunc("/websocket", restcore.Authenticator(RunWebsocket, db)).Headers("Upgrade", "websocket").Methods("GET")

	crud.Router(db, prefix.PathPrefix("/crud").Subrouter())
	query.Router(db, prefix.PathPrefix("/query").Subrouter())
	feed.Router(db, prefix.PathPrefix("/feed").Subrouter())
	meta.Router(db, prefix.PathPrefix("/meta").Subrouter())

	//login and Logout of the system
	prefix.HandleFunc("/login", restcore.Authenticator(Login, db)).Methods("GET")
	prefix.HandleFunc("/logout", restcore.Authenticator(Logout, db)).Methods("GET")

	//Now that things are running, we want the ability to do a clean shutdown of REST
	util.CloseOnExit(restcloser{})

	return prefix, nil
}
