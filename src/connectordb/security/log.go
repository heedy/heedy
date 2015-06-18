package security

import (
	"net/http"
	"net/http/httputil"

	log "github.com/Sirupsen/logrus"
)

/** LoggingHandler provdies logging information for every request that goes
through a system.
**/
func LoggingHandler(h http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		dump, _ := httputil.DumpRequest(request, true)
		log.Infof("HTTP Request (bytes): %v", string(dump))
		h.ServeHTTP(writer, request)
	})
}
