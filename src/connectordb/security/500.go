package security

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
)

/** Catches 500 errors and logs them to the display with a UUID while returning
a nice result for the user.
**/
func FiveHundredHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writerWrapper := NewErrorMaskingResponseWriter(writer)
		h.ServeHTTP(&writerWrapper, request)
	})
}

// NewErrorMaskingResponseWriter creates an error masking response writer that logs
// 500 errors to the log files and returns a reference number to the user so we hide
// internal errors
func NewErrorMaskingResponseWriter(writer http.ResponseWriter) ErrorMaskingResponseWriter {
	u, _ := uuid.NewV4()
	uu := u.String()
	return ErrorMaskingResponseWriter{ResponseWriter: writer, responseCode: 200, uuid: uu}
}

type ErrorMaskingResponseWriter struct {
	http.ResponseWriter
	responseCode  int
	uuid          string
	shouldCapture bool
}

// basic wrapper
func (e *ErrorMaskingResponseWriter) Header() http.Header {
	return e.ResponseWriter.Header()
}

// Catch 500 errors and supress them
func (e *ErrorMaskingResponseWriter) WriteHeader(responseCode int) {
	e.responseCode = responseCode
	e.ResponseWriter.WriteHeader(responseCode)

	if e.responseCode >= 500 && e.responseCode < 600 {
		e.ResponseWriter.Write([]byte("Oops, something messed up, if you have questions contact us with the id: " + e.uuid))
		e.shouldCapture = true
	}
}

func (e *ErrorMaskingResponseWriter) Write(body []byte) (int, error) {
	if e.shouldCapture {
		fields := log.Fields{"uuid": e.uuid, "response code": e.responseCode}
		logger := log.WithFields(fields)
		logger.Warningf(string(body))

		// we "wrote" the whole thing
		return len(body), nil
	} else {
		return e.ResponseWriter.Write(body)
	}
}
