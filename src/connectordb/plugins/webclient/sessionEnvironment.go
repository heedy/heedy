package webclient

import (
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"
	"errors"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
)

var (
	store               = sessions.NewCookieStore([]byte("web-service-special-key"))
	sessionName         = "connectordb_login"
	ErrUserDevNotStored = errors.New("User or device could not be found.")
)

// The environment for a particular session
type SessionEnvironment struct {
	Writer   http.ResponseWriter
	Request  *http.Request
	Session  *sessions.Session
	User     *users.User
	Device   *users.Device
	Operator operator.Operator
}

// Logs a user out of the system by destroying keys in their session
func (se *SessionEnvironment) Logoff() {
	se.Session.Values["User"] = nil
	se.Session.Values["Device"] = nil
}

// Saves the session environment
func (se *SessionEnvironment) Save() {
	store.Save(se.Request, se.Writer, se.Session)
}

// HandleError handles a given error if it exists, if the error was caught and
// logged, we add a flash, and return true.
func (se *SessionEnvironment) HandleError(err error, flash string, logger *log.Entry) bool {
	if err == nil {
		return false
	}

	logger.Warn(err.Error())
	se.Session.AddFlash(flash)

	return true
}

func NewSessionEnvironment(rw http.ResponseWriter, req *http.Request) (se SessionEnvironment, err error) {
	se.Writer = rw
	se.Request = req
	se.Session, err = store.Get(req, sessionName)
	if err != nil {
		return se, err
	}

	usr, ok := se.Session.Values["User"]
	if !ok || usr == nil {
		return se, ErrUserDevNotStored
	}
	u := usr.(users.User)
	se.User = &u

	dev, ok := se.Session.Values["Device"]
	if !ok || dev == nil {
		return se, ErrUserDevNotStored
	}
	d := dev.(users.Device)
	se.Device = &d

	se.Operator, err = userdb.DeviceOperator(se.Device.DeviceId)
	return se, err
}
