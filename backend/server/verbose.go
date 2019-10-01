package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"

	"github.com/sirupsen/logrus"
)

// VerboseLoggingMiddleware performs extremely verbose logging - including all incoming requests and responses.
// This can be activated using --vvv on the server
func VerboseLoggingMiddleware(h http.Handler, log *logrus.Entry) http.Handler {
	if log == nil {
		log = logrus.NewEntry(logrus.StandardLogger())
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		req, err := httputil.DumpRequest(request, true)
		if err != nil {
			log.Error(err)
			http.Error(writer, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Debugf("Request:\n\n%s\n\n", string(req))

		// We don't want to mess with websocket connections
		if request.Header.Get("Connection") == "Upgrade" {
			h.ServeHTTP(writer, request)
			return
		}

		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, request)

		response := rec.Body.Bytes()

		headers := ""
		for k, v := range rec.HeaderMap {
			curheader := k + ":"
			for s := range v {
				curheader += " " + v[s]
			}
			headers += curheader + "\n"
		}

		if v, ok := rec.HeaderMap["Content-Encoding"]; ok && len(v) > 0 && v[0] != "identity" {
			log.Debugf("Response: %d\n\n%s\n\nRESPONSE BODY COMPRESSED - NOT LOGGING (length: %d)", rec.Code, headers, len(response))
		} else if ctype := rec.Header().Get("Content-Type"); strings.HasPrefix(ctype, "font/") || strings.HasPrefix(ctype, "image/") || strings.HasPrefix(ctype, "video/") || strings.HasPrefix(ctype, "audio/") || ctype == "application/octet-stream" || ctype == "application/pdf" || ctype == "application/zip" {
			log.Debugf("Response: %d\n\n%s\n\nBINARY CONTENT-TYPE - NOT LOGGING (length: %d)", rec.Code, headers, len(response))
		} else {
			// http://stackoverflow.com/questions/27983893/in-go-how-to-inspect-the-http-response-that-is-written-to-http-responsewriter
			log.Debugf("Response: %d\n\n%s\n%s\n\n", rec.Code, headers, string(response))
		}

		// Now copy everything from response recorder to actual response writer
		// http://stackoverflow.com/questions/29319783/go-logging-responses-to-incoming-http-requests-inside-http-handlefunc
		for k, v := range rec.HeaderMap {
			writer.Header()[k] = v
		}
		writer.WriteHeader(rec.Code)
		writer.Write(response)

	})
}
