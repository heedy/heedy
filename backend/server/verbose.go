package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"unicode/utf8"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/sirupsen/logrus"
)

// VerboseLoggingMiddleware performs extremely verbose logging - including all incoming requests and responses.
// This can be activated using --vvv on the server
func VerboseLoggingMiddleware(h http.Handler, log *logrus.Entry) http.Handler {
	if log == nil {
		log = logrus.NewEntry(logrus.StandardLogger())
	}

	VerboseLogBufferLength := 1024
	vbl := assets.Config().VerboseLogBuffer
	if vbl != nil && *vbl > 4*3 {
		VerboseLogBufferLength = *vbl
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		req, err := httputil.DumpRequest(request, true)
		if err != nil {
			log.Error(err)
			http.Error(writer, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		reqs := string(req)
		if len(reqs) > VerboseLogBufferLength {
			reqs = reqs[:VerboseLogBufferLength] + "\n(REQUEST LOG TRUNCATED)"
		}
		log.Debugf("Request:\n\n%s\n\n", reqs)

		// We don't want to mess with websocket apps
		if request.Header.Get("Upgrade") == "websocket" {
			log.Debug("Websocket request - not logging raw response")
			h.ServeHTTP(writer, request)
			return
		}

		rs := run.NewResponseStreamer()
		rs.Serve(h, request)

		buf := make([]byte, VerboseLogBufferLength)

		//Read the first part of the request into the buffer
		n, err := io.ReadFull(rs, buf)

		reqstring := fmt.Sprintf("%s %s", request.Method, request.URL.Path)

		go func() {
			headers := ""
			for k, v := range rs.Header() {
				curheader := k + ":"
				for s := range v {
					curheader += " " + v[s]
				}
				headers += curheader + "\n"
			}
			response := string(buf[:n])
			if n == 0 {
				log.Debugf("Response: %d (%s)\n\n%s\n\n", rs.Code, reqstring, headers)
			} else if v, ok := rs.Header()["Content-Encoding"]; ok && len(v) > 0 && v[0] != "identity" {
				log.Debugf("Response: %d (%s)\n\n%s\n\nRESPONSE BODY COMPRESSED - NOT LOGGING\n\n", rs.Code, reqstring, headers)
			} else if ctype := rs.Header().Get("Content-Type"); strings.HasPrefix(ctype, "font/") || strings.HasPrefix(ctype, "image/") || strings.HasPrefix(ctype, "video/") || strings.HasPrefix(ctype, "audio/") || ctype == "application/octet-timeseries" || ctype == "application/pdf" || ctype == "application/zip" {
				log.Debugf("Response: %d (%s)\n\n%s\n\nBINARY CONTENT-TYPE - NOT LOGGING\n\n", rs.Code, reqstring, headers)
			} else {
				if n < VerboseLogBufferLength {
					log.Debugf("Response: %d (%s)\n\n%s\n%s\n\n", rs.Code, reqstring, headers, response)
				} else {
					// If it isn't utf8 valid, we probably stopped read in the middle of a utf8 character, so remove the last 3 code points
					// just in case.
					if !utf8.ValidString(response) {
						runeres := []rune(response)
						response = string(runeres[:len(runeres)-3])
					}
					log.Debugf("Response: %d (%s)\n\n%s\n%s\n(RESPONSE LOG TRUNCATED)\n\n", rs.Code, reqstring, headers, response)
				}
			}
		}()

		// Now copy everything from response recorder to actual response writer
		// http://stackoverflow.com/questions/29319783/go-logging-responses-to-incoming-http-requests-inside-http-handlefunc
		for k, v := range rs.Header() {
			writer.Header()[k] = v
		}
		writer.WriteHeader(rs.Code)
		writer.Write(buf[:n])
		if err == nil {
			io.Copy(writer, rs)
		}

		rs.Close()
	})
}
