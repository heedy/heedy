package webclient

import (
	"net/http"
	"streamdb/users"
	"github.com/gorilla/sessions"
	"errors"
)

var (
	DecodeError = errors.New("Session decoding error")
)


// Session response writer saves a session before headers are written
type SessionResponseWriter struct {
	rw http.ResponseWriter
	req *http.Request
	session *sessions.Session
}

// Creates a new session response writer
func NewSessionResponseWriter(rw http.ResponseWriter, req *http.Request, session *sessions.Session) *SessionResponseWriter {
	return &SessionResponseWriter{rw, req, session}
}


// Get the underlying request
func (srw *SessionResponseWriter) Request() *http.Request {
	return srw.req
}

// Get the underlying session
func (srw *SessionResponseWriter) Session() *sessions.Session {
	return srw.session
}

// Get the user and the user device from this session
func (srw *SessionResponseWriter) GetUserAndDevice() (*users.User, *users.Device, error) {
	user, ok := srw.session.Values["User"].(users.User)

	if ! ok {
		return nil, nil, DecodeError
	}

	device, ok := srw.session.Values["Device"].(users.Device)

	if ! ok {
		return nil, nil, DecodeError
	}

	return &user, &device, nil
}

// Header returns the header map that will be sent by WriteHeader.
// Changing the header after a call to WriteHeader (or Write) has
// no effect.
func (srw *SessionResponseWriter) Header() http.Header {
	return srw.rw.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
// before writing the data.  If the Header does not contain a
// Content-Type line, Write adds a Content-Type set to the result of passing
// the initial 512 bytes of written data to DetectContentType.
func (srw *SessionResponseWriter) Write(b []byte) (int, error) {
	srw.session.Save(srw.req, srw.rw)
	return srw.rw.Write(b)
}

// WriteHeader sends an HTTP response header with status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (srw *SessionResponseWriter) WriteHeader(i int) {
	srw.rw.WriteHeader(i)
}
