package server

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"

	"github.com/heedy/heedy/api/golang/rest"
)

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// Oauth errors are of a specific format that we will follow, just to make sure we never break it:
// https://www.oauth.com/oauth2-servers/access-tokens/access-token-response/
func writeAuthError(w http.ResponseWriter, r *http.Request, status int, errVal, errDescription string) {
	c := rest.CTX(r)

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
	DB *database.AdminDB

	codeCache *cache.Cache
}

// NewAuth creates a new oauth flow handler using an admin DB
func NewAuth(db *database.AdminDB) *Auth {
	return &Auth{
		DB:        db,
		codeCache: cache.New(5*time.Minute, 5*time.Minute),
	}
}

func (a *Auth) Authenticate(w http.ResponseWriter, r *http.Request) (database.DB, error) {
	// First, try authenticating as a app
	accessToken := r.Header.Get("Authorization")
	if len(accessToken) > 0 {
		const prefix = "Bearer "
		if len(accessToken) < len(prefix) || !strings.EqualFold(accessToken[:len(prefix)], prefix) {
			return nil, errors.New("bad_request: Malformed authorization header")
		}
		accessToken = accessToken[len(prefix):]
	} else {
		// No authorization header. Check the url params for a token
		accessToken = r.URL.Query().Get("access_token")
	}

	if len(accessToken) > 0 {
		// Try logging in as a app
		c, err := a.DB.GetAppByAccessToken(accessToken)
		if err != nil {
			return nil, errors.New("access_denied: invalid API key")
		}
		if !*c.Enabled {
			return nil, errors.New("app_disabled: the app was disabled")
		}
		return database.NewAppDB(a.DB, c), nil

	}

	// Then see if there was a cookie
	cookie, err := r.Cookie("token")
	if err == nil {
		// There was a cookie. Cookie errors are just treated
		// as if the auth didn't exist.

		if cookie.Name == "token" && cookie.Value != "" {
			username, _, err := a.DB.GetUserSessionByToken(cookie.Value) // user name not currently used
			if err == nil {
				// Return the logged in user database
				return database.NewUserDB(a.DB, username), nil
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    "",
				MaxAge:   -1,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})
		}
	}

	// Nobody is logged in, return a public database view
	return database.NewPublicDB(a.DB), nil
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
		uname, _, err := a.DB.AuthUser(usr, password)
		if err != nil {
			time.Sleep(1 * time.Second) // Wait a second before returning failure
			writeAuthError(w, r, 400, "access_denied", "Wrong username or password")
			return
		}
		// Add the token
		tok, _, err := a.DB.CreateUserSession(uname, r.Header.Get("User-Agent"))
		if err != nil {
			writeAuthError(w, r, 400, "server_error", err.Error())
			return
		}

		// Set a cookie - technically the password grant should return json,
		// but we will actually set the cookie anyways, so we directly get whether
		// the user is logged in with each request
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tok,
			Expires:  time.Now().AddDate(5, 0, 0),
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			HttpOnly: true,
		})

		// ... and also return the json response
		rest.WriteJSON(w, r, &tokenResponse{
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
// create a app. It is identical to a standard oauth request authorization
// code request if the client is known. If it is an unknown client,
// allows the client to request creation of a specific app on its behalf.
type CodeRequest struct {
	// These are parameters of an authorization request on Oauth2
	// https://www.oauth.com/oauth2-servers/authorization/the-authorization-request/
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	State       string `json:"state,omitempty"`
	Scope       string `json:"scope,omitempty"`

	// The app object to create - if clientID is not set
	App *database.App
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
		ctx := rest.CTX(r)
		ctx.Log.Debug("Running auth template")
		aTemplate.Execute(w, &aContext{
			User:    nil,
			Request: nil,
		})
		return
	})

	mux.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		c := rest.CTX(r)

		// Remove all site data
		w.Header().Add("Clear-Site-Data", "\"*\"")

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    "",
			MaxAge:   0,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		})

		// Should verify that getting correct referrer
		http.Redirect(w, r, "/", 303)

		// We use the happy path - this never fails. At worst it
		v, err := r.Cookie("token")
		if err != nil {
			c.Log.Error(err)
			return
		}
		if v.Value != "" {
			err = c.DB.AdminDB().DelUserSessionByToken(v.Value)
			if err != nil {
				c.Log.Error(err)
			}
		}
	})
	return mux, nil
}
