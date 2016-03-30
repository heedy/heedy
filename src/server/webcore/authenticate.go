/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package webcore

import (
	"connectordb"
	"connectordb/authoperator"
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
func Authenticate(db *connectordb.Database, request *http.Request) (o *authoperator.AuthOperator, err error) {
	//Basic auth overrides all other auth
	authUser, authPass, ok := request.BasicAuth()

	if ok {
		if authUser != "" {
			o, err = db.UserLogin(authUser, authPass)

		} else {
			o, err = db.DeviceLogin(authPass)
		}
	} else {
		//Basic auth is unavailable.

		//Check if there is an apikey parameter in the query itself
		authPass = request.URL.Query().Get("apikey")
		if len(authPass) != 0 {
			o, err = db.DeviceLogin(authPass)
		} else {
			var cookie *http.Cookie
			cookie, err = request.Cookie("connectordb-session")
			if err == nil {
				err = CookieMonster.Decode("connectordb-session", cookie.Value, &authPass)
				if err == nil {
					o, err = db.DeviceLogin(authPass)
				} else {
					//If the cookie is invalid, log in as nobody
					err = nil
					o = db.Nobody()
				}
			} else {
				err = nil
				// No authentication was given - use nobody
				o = db.Nobody()
			}
		}
	}

	if err != nil {
		atomic.AddUint32(&StatsAuthFails, 1)
	}
	return o, err
}

//CreateSessionCookie generates the authentication cookie from an authenticated user
func CreateSessionCookie(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request) error {
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

	encoded, err := CookieMonster.Encode("connectordb-session", dev.APIKey)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:  "connectordb-session",
		Value: encoded,
		Path:  "/",
	}

	//Check a remember me param
	if request != nil {
		val, ok := request.URL.Query()["remember"]
		if ok && val[0] == "true" {
			cookie.MaxAge = CookieMaxAge
		}
	}

	http.SetCookie(writer, cookie)

	return nil
}

// HasSession returns whether there is a session cookie with this request
func HasSession(request *http.Request) bool {
	_, err := request.Cookie("connectordb-session")
	return err == nil
}
