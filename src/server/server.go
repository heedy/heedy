/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package server

import (
	"config"
	"connectordb"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"server/restapi"
	"server/restapi/restcore"
	"server/webcore"
	"server/website"
	"strings"
	"sync/atomic"

	"github.com/dkumor/acmewrapper"
	"github.com/gorilla/mux"
	"github.com/xenolf/lego/acme"

	stdlog "log"

	log "github.com/Sirupsen/logrus"
)

//SecurityHeaderHandler provides a wrapper function for an http.Handler that sets several security headers for all sessions passing through
func SecurityHeaderHandler(h http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// See the OWASP security project for these headers:
		// https://www.owasp.org/index.php/List_of_useful_HTTP_headers

		// Don't allow our site to be embedded in another
		writer.Header().Set("X-Frame-Options", "deny")

		// Enable the client side XSS filter
		writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// Disable content sniffing which could lead to improperly executed
		// scripts or such from malicious user uploads
		writer.Header().Set("X-Content-Type-Options", "nosniff")

		h.ServeHTTP(writer, request)
	})
}

//OptionsHandler on OPTIONS to allow cross-site XMLHTTPRequest, allow access control origin
func OptionsHandler(writer http.ResponseWriter, request *http.Request) {
	webcore.GetRequestLogger(request, "OPTIONS").Debug()
	webcore.WriteAccessControlHeaders(writer, request)

	//These headers are only needed for the OPTIONS request
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	writer.WriteHeader(http.StatusOK)
}

// NotFoundHandler handles 404 errors for the whole server
func NotFoundHandler(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/api") {
		logger := webcore.GetRequestLogger(request, "404")
		//If this is a REST API call, write a REST-like error
		atomic.AddUint32(&webcore.StatsRESTQueries, 1)
		restcore.WriteError(writer, logger, http.StatusNotFound, errors.New("This path is not recognized"), false)
		return
	}

	//Otherwise, we assume that it is the web not found handler
	website.NotFoundHandler(writer, request)

}

// Redirect80 Redirects port 80 to the given site url
func Redirect80(siteURL string) {
	log.Error(http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, siteURL+r.RequestURI, http.StatusMovedPermanently)
	})))
}

//RunServer runs the ConnectorDB frontend server
func RunServer() error {
	OSSpecificSetup()

	// ACME has a special logger, so set it
	acme.Logger = stdlog.New(log.StandardLogger().Writer(), "", 0)
	acmewrapper.Logger = log.StandardLogger()

	// Gets the global server configuration
	c := config.Get()

	err := webcore.Initialize(c)
	if err != nil {
		return err
	}
	// Reload webcore settings on config change
	config.OnChangeCallback(webcore.Initialize)

	//Connect using the configuration
	db, err := connectordb.Open(c.Options())
	if err != nil {
		return err
	}

	r := mux.NewRouter()

	//Allow for the application to match /path and /path/ to the same place.
	r.StrictSlash(true)

	//Setup the 404 handler
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	r.Methods("OPTIONS").Handler(http.HandlerFunc(OptionsHandler))

	//The rest api has its own versioned url
	s := r.PathPrefix("/api/v1").Subrouter()
	_, err = restapi.Router(db, s)
	if err != nil {
		return err
	}

	//The website is initialized at /
	_, err = website.Router(db, r)
	if err != nil {
		return err
	}

	//Set up the web server
	handler := SecurityHeaderHandler(r)

	//Show the statistics
	go webcore.RunStats()
	go webcore.RunQueryTimers()

	//Run the dbwriter
	go db.RunWriter()

	if c.Redirect80 {
		go Redirect80(c.GetSiteURL())
	}

	listenhost := fmt.Sprintf("%s:%d", c.Hostname, c.Port)

	//Run an https server if we are given tls cert and key
	if c.TLSEnabled() {
		if c.TLS.ACME.Enabled {
			log.Debugf("Attempting to use ACME with host %s", listenhost)
		}
		// Enable http2 support &Let's Encrypt support
		w, err := acmewrapper.New(acmewrapper.Config{
			Address:          listenhost,
			Server:           c.TLS.ACME.Server,
			PrivateKeyFile:   c.TLS.ACME.PrivateKey,
			RegistrationFile: c.TLS.ACME.Registration,
			Domains:          c.TLS.ACME.Domains,
			TLSCertFile:      c.TLS.Cert,
			TLSKeyFile:       c.TLS.Key,
			TOSCallback:      acmewrapper.TOSAgree,
			AcmeDisabled:     !c.TLS.ACME.Enabled,
		})
		if err != nil {
			return err
		}
		tlsconfig := w.TLSConfig()

		listener, err := tls.Listen("tcp", listenhost, tlsconfig)
		if err != nil {
			return err
		}

		server := &http.Server{
			Addr:      listenhost,
			Handler:   handler,
			TLSConfig: tlsconfig,
		}
		acmestring := ""
		if c.TLS.ACME.Enabled {
			acmestring = " ACME"
		}

		log.Infof("Running ConnectorDB v%s at %s (%s TLS%s)", connectordb.Version, c.GetSiteURL(), listenhost, acmestring)

		return server.Serve(listener)
	}
	log.Infof("Running ConnectorDB v%s at %s (%s)", connectordb.Version, c.GetSiteURL(), listenhost)
	http.Handle("/", handler)

	return http.ListenAndServe(listenhost, nil)
}
