package server

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
)

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// Oauth errors are of a specific format that we will follow, just to make sure we never break it:
// https://www.oauth.com/oauth2-servers/access-tokens/access-token-response/
func writeAuthError(w http.ResponseWriter, r *http.Request, status int, errVal, errDescription string) {
	c := CTX(r)

	er := oauthErrorResponse{
		Error:            errVal,
		ErrorDescription: errDescription,
	}
	jer, err := json.Marshal(&er)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "server_error", "error_description": "Failed to create error message"}`))
		if c != nil {
			c.Log.Errorf("Failed to write error message: %s", err)
		} else {
			logrus.Errorf("Failed to write error message: %s", err)
		}
	}

	if c != nil {
		c.Log.Warn(errVal)
	} else {
		logrus.Warn(errVal)
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jer)))
	w.WriteHeader(status)
	w.Write(jer)
}

// Auth handles the oauth flow
type Auth struct {
	db *database.AdminDB

	codeCache *cache.Cache
}

// NewAuth creates a new oauth flow handler using an admin DB
func NewAuth(db *database.AdminDB) *Auth {
	return &Auth{
		db:        db,
		codeCache: cache.New(5*time.Minute, 5*time.Minute),
	}
}

func (a *Auth) Authenticate(r *http.Request) (database.DB, error) {
	cookie, err := r.Cookie("token")
	if err == nil {
		// There was a cookie. Cookie errors are just treated
		// as if the auth didn't exist.

		if cookie.Name == "token" && cookie.Value != "" {
			_, err := a.db.LoginToken(cookie.Value) // user name not currently used
			if err == nil {
				// Return the logged in user database
				return a.db, nil
			}
		}
		r.AddCookie(&http.Cookie{
			Name:   cookie.Name,
			Value:  "",
			MaxAge: 0,
		})
	}

	// Nobody is logged in, return a public database view
	return a.db, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
	State       string `json:"state,omitempty"`
}

// ServeToken handles a post request to the token endpoint.
// It handles password grants and authorization code requests
func (a *Auth) ServeToken(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		writeAuthError(w, r, 400, "invalid_request", "Could not parse request")
		return
	}
	switch grant := r.FormValue("grant_type"); grant {
	case "password":
		// The grant can only be requested from current
		// The password grant type must have a valid username password combo
		usr := r.FormValue("username")
		password := r.FormValue("password")
		if usr == "" || password == "" {
			writeAuthError(w, r, 400, "parameter_absent", "Must have both username and password")
			return
		}
		uname, _, err := a.db.AuthUser(usr, password)
		if err != nil {
			writeAuthError(w, r, 400, "access_denied", "No user was found with the given password")
			return
		}
		// Add the token
		tok, err := a.db.AddLoginToken(uname)
		if err != nil {
			writeAuthError(w, r, 400, "server_error", err.Error())
			return
		}

		// Set a cookie - technically the password grant should return json,
		// but we will actually set the cookie anyways, so we directly get whether
		// the user is logged in with each request
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tok,
			Expires: time.Now().AddDate(5, 0, 0),
		})

		// ... and also return the json response
		WriteJSON(w, r, &tokenResponse{
			AccessToken: tok,
			TokenType:   "bearer",
		}, nil)

	default:
		writeAuthError(w, r, 400, "unsupported_grant_type", "Grant type not supported")
		return
	}

}

// ServeCode handles a post request to the code endpoint
func (a *Auth) ServeCode(w http.ResponseWriter, r *http.Request) {

}

// CodeRequest is sent in by the client trying to
// create a connection. It is identical to a standard oauth request authorization
// code request if the client is known. If it is an unknown client,
// allows the client to request creation of a specific connection on its behalf.
type CodeRequest struct {
	// These are parameters of an authorization request on Oauth2
	// https://www.oauth.com/oauth2-servers/authorization/the-authorization-request/
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	State       string `json:"state,omitempty"`
	Scope       string `json:"scope,omitempty"`

	// The connection object to create - if clientID is not set
	Connection *database.Connection
}

// RequestCode returns the information relevant to an authorization code request
func (a *Auth) RequestCode(r *http.Request) (*CodeRequest, error) {
	return nil, errors.New("Not implemented")
}

func AuthMux(a *Auth) (*chi.Mux, error) {
	mux := chi.NewMux()

	// The authorization flow (login/give permissions page)
	abytes, err := afero.ReadFile(assets.Get().FS, "/public/auth.html")
	if err != nil {
		return nil, err
	}
	aTemplate, err := template.New("auth").Parse(string(abytes))
	if err != nil {
		return nil, err
	}
	mux.Post("/token", a.ServeToken)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {

		// Disallow clickjacking
		// https://www.oauth.com/oauth2-servers/authorization/security-considerations/
		w.Header().Add("X-Frame-Options", "DENY")
		ctx := CTX(r)
		ctx.Log.Debug("Running auth template")
		aTemplate.Execute(w, &aContext{
			User:    nil,
			Request: nil,
		})
		return
	})
	return mux, nil
}
