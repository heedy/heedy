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
	"net/http/httptest"
	"net/http/httputil"
	"server/restapi"
	"server/restapi/restcore"
	"server/webcore"
	"server/website"
	"strings"
	"sync/atomic"
	"time"

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

// VerboseLoggingHandler performs extremely verbose logging - including all incoming requests and responses.
// This can be activated using --vvv on the server
func VerboseLoggingHandler(h http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		logger := webcore.GetRequestLogger(request, "VERBOSE")

		// We don't want to mess with websocket connections
		if request.Header.Get("Upgrade") == "WebSocket" {
			logger.Warn("Can't log websocket connections in verbose mode")
			h.ServeHTTP(writer, request)
			return
		}

		req, err := httputil.DumpRequest(request, true)
		if err != nil {
			logger.Error(err)
			http.Error(writer, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		logger.WithField("type", "REQUEST").Debugf("Request:\n\n%s\n\n", string(req))

		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, request)

		// http://stackoverflow.com/questions/27983893/in-go-how-to-inspect-the-http-response-that-is-written-to-http-responsewriter
		response := rec.Body.Bytes()
		logger.WithField("type", "RESPONSE").Debugf("Response: %d\n\n%s\n\n", rec.Code, string(response))

		// Now copy everything from response recorder to actual response writer
		// http://stackoverflow.com/questions/29319783/go-logging-responses-to-incoming-http-requests-inside-http-handlefunc
		for k, v := range rec.HeaderMap {
			writer.Header()[k] = v
		}
		writer.WriteHeader(rec.Code)
		writer.Write(response)

	})
}

// MakeHandler generates the handler for the server. It adds the verbose middleware if it is needed
func MakeHandler(h http.Handler, verbose bool) http.Handler {
	if verbose {
		return VerboseLoggingHandler(h)
	}
	return h
}

//RunServer runs the ConnectorDB frontend server
func RunServer(verbose bool) error {
	OSSpecificSetup()

	if verbose {
		log.Warn("Running in verbose mode. Use this for debugging only!")
	}

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
	r.NotFoundHandler = MakeHandler(http.HandlerFunc(NotFoundHandler), verbose)

	r.Methods("OPTIONS").Handler(MakeHandler(http.HandlerFunc(OptionsHandler), verbose))

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
	handler := MakeHandler(SecurityHeaderHandler(r), verbose)

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
			Addr:        listenhost,
			Handler:     handler,
			TLSConfig:   tlsconfig,
			ReadTimeout: time.Duration(c.HTTPReadTimeout) * time.Second,
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
