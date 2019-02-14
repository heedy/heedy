/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package website

//Router returns a fully formed Gorilla router given an optional prefix
import (
	"compress/gzip"
	"config"
	"connectordb"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
)

// Screw it, let's just use a global database
var Database *connectordb.Database

// Handler middleware for statically served files
// Set up both caching and gzip,since these files are public
// https://gist.github.com/bryfry/09a650eb8aac0fb76c24
// https://play.golang.org/p/fpETA9_1oo

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func staticFileHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()
		if cfg.CacheStatic {
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age:%d, public", cfg.CacheStaticAge))
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || !cfg.GzipStatic {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		h.ServeHTTP(gzw, r)
	})
}

func specificFileHandler(filename string) {

}

// Router handles the website
func Router(db *connectordb.Database, r *mux.Router) (*mux.Router, error) {
	if r == nil {
		r = mux.NewRouter()
	}

	Database = db

	err := LoadFiles()
	if err != nil {
		return nil, err
	}

	//Allow for the application to match /path and /path/ to the same place.
	r.StrictSlash(true)

	//The app and web prefixes are served directly from the correct directories
	www := "/" + WWWPrefix
	r.PathPrefix(www).Handler(http.StripPrefix(www, staticFileHandler(http.FileServer(http.Dir(WWWPath)))))
	app := "/" + AppPrefix
	r.PathPrefix(app).Handler(http.StripPrefix(app, staticFileHandler(http.FileServer(http.Dir(AppPath)))))

	//Handle the favicon
	r.Handle("/favicon.ico", http.RedirectHandler(www+"/favicon.ico", http.StatusMovedPermanently))

	// Robots and sitemap
	r.Handle("/robots.txt", http.RedirectHandler(www+"/robots.txt", http.StatusMovedPermanently))
	r.Handle("/sitemap.xml", http.RedirectHandler(www+"/sitemap.xml", http.StatusMovedPermanently))

	// ServiceWorker needs to be at root of domain to handle all requests. The serviceworker js is assumed
	// to be in /app/serviceworker.js.
	// Unfortunately, chrome doesn't allow serviceworkers to use redirects... So we must specifically handle this file manually
	r.Handle("/serviceworker.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(AppPath, "serviceworker.js"))
	}))

	// These functions are actually fairly standard for login/logout across different apps
	// so we make them work the same way here
	r.Handle("/logout", http.HandlerFunc(LogoutHandler)).Methods("GET")
	r.Handle("/login", Authenticator(WWWLogin, Login, db)).Methods("GET")

	// Handle creation of new users
	r.Handle("/join", http.HandlerFunc(JoinHandleGET)).Methods("GET")
	r.Handle("/join", http.HandlerFunc(JoinHandlePOST)).Methods("POST")

	//Now load the user/device/stream paths
	r.HandleFunc("/", Authenticator(WWWIndex, Index, db)).Methods("GET")
	r.HandleFunc("/{user}", Authenticator(WWWLogin, User, db)).Methods("GET")
	r.HandleFunc("/{user}/{device}", Authenticator(WWWLogin, Device, db)).Methods("GET")
	r.HandleFunc("/{user}/{device}/{stream}", Authenticator(WWWLogin, Stream, db)).Methods("GET")

	return r, nil
}
