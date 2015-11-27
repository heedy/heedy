package webcore

import (
	"connectordb"
	"connectordb/operator"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/securecookie"
)

var (
	//UnsuccessfulLoginWait is the amount of time to wait between each unsuccessful login attempt
	UnsuccessfulLoginWait = 300 * time.Millisecond

	// ErrNoAuthentication is an error that is thrown when no authentication is given
	ErrNoAuthentication = errors.New("No authentication given with request")

	//CookieMonster is the handler for authentication based on cookies
	CookieMonster *securecookie.SecureCookie

	//The time in seconds to keep a cookie valid
	CookieMaxAge = 60 * 60 * 24 * 30 * 4
)

// Authenticate gets the authenticated device Operator given an http.Request
func Authenticate(db *connectordb.Database, request *http.Request) (operator.Operator, error) {

	//Basic auth overrides all other auth
	authUser, authPass, ok := request.BasicAuth()

	if !ok {
		//Basic auth is unavailable.

		//Check if there is an apikey parameter in the query itself
		authPass = request.URL.Query().Get("apikey")

		//If there was no apikey, check for a cookie
		if len(authPass) == 0 {
			cookie, err := request.Cookie("connectordb-session")
			if err != nil {
				atomic.AddUint32(&StatsAuthFails, 1)
				return nil, ErrNoAuthentication
			}
			if err = CookieMonster.Decode("connectordb-session", cookie.Value, &authPass); err != nil {
				atomic.AddUint32(&StatsAuthFails, 1)
				return nil, err
			}
		}
	}

	//If we got here, it looks like some form of auth was extracted.
	o, err := operator.NewPathLoginOperator(db, authUser, authPass)

	if err != nil {
		atomic.AddUint32(&StatsAuthFails, 1)
		time.Sleep(UnsuccessfulLoginWait)
	}
	return o, err
}

//CreateSessionCookie generates the authentication cookie from an authenticated user
func CreateSessionCookie(o operator.Operator, writer http.ResponseWriter, remember bool) error {
	if o == nil {
		//If the operator is nil, we delete the cookie
		cookie := &http.Cookie{
			Name:   "connectordb-session",
			MaxAge: -1,
			Path:   "/",
		}
		http.SetCookie(writer, cookie)
		return nil
	}
	dev, err := o.Device()
	if err != nil {
		return err
	}

	encoded, err := CookieMonster.Encode("connectordb-session", dev.ApiKey)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:  "connectordb-session",
		Value: encoded,
		Path:  "/",
	}

	//Check for a "remember" parameter in cookie
	if remember {
		cookie.MaxAge = CookieMaxAge
	}

	http.SetCookie(writer, cookie)

	return nil
}

// HasSession returns whether there is a session cookie with this request
func HasSession(request *http.Request) bool {
	_, err := request.Cookie("connectordb-session")
	return err == nil
}
