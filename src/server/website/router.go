/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

//Router returns a fully formed Gorilla router given an optional prefix
import (
	"connectordb"
	"net/http"

	"github.com/gorilla/mux"
)

// Router handles the website
func Router(db *connectordb.Database, r *mux.Router) (*mux.Router, error) {
	if r == nil {
		r = mux.NewRouter()
	}

	err := LoadFiles()
	if err != nil {
		return nil, err
	}

	//Allow for the application to match /path and /path/ to the same place.
	r.StrictSlash(true)

	//The app and web prefixes are served directly from the correct directories
	www := "/" + WWWPrefix
	r.PathPrefix(www).Handler(http.StripPrefix(www, http.FileServer(http.Dir(WWWPath))))
	app := "/" + AppPrefix
	r.PathPrefix(app).Handler(http.StripPrefix(app, http.FileServer(http.Dir(AppPath))))

	//Handle the favicon
	r.Handle("/favicon.ico", http.RedirectHandler(www+"/favicon.ico", http.StatusMovedPermanently))
	r.Handle("/robots.txt", http.RedirectHandler(www+"/robots.txt", http.StatusMovedPermanently))
	r.Handle("/sitemap.xml", http.RedirectHandler(www+"/sitemap.xml", http.StatusMovedPermanently))

	// These functions are actually fairly standard for login/logout across different apps
	// so we make them work the same way here
	r.Handle("/logout", http.HandlerFunc(LogoutHandler)).Methods("GET")
	r.Handle("/login", Authenticator(WWWLogin, Login, db)).Methods("GET")

	r.Handle("/join", http.HandlerFunc(JoinHandler)).Methods("GET")

	//Now load the user/device/stream paths
	r.HandleFunc("/", Authenticator(WWWIndex, Index, db)).Methods("GET")
	r.HandleFunc("/{user}", Authenticator(WWWLogin, User, db)).Methods("GET")
	r.HandleFunc("/{user}/{device}", Authenticator(WWWLogin, Device, db)).Methods("GET")
	r.HandleFunc("/{user}/{device}/{stream}", Authenticator(WWWLogin, Stream, db)).Methods("GET")

	return r, nil
}
