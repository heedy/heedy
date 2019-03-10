package auth

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/database"
	"github.com/patrickmn/go-cache"
)

// Auth codes
var codeCache = cache.New(5*time.Minute, 5*time.Minute)

// AuthRequest is sent in by the client trying to
// create a connection.
type AuthRequest struct {
	// These are parameters of an authorization request on Oauth2
	// https://www.oauth.com/oauth2-servers/authorization/the-authorization-request/
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	State       string `json:"state,omitempty"`
	Scope       string `json:"scope,omitempty"`

	// The connection object to create - if clientID is not set
	Connection *database.Connection
}

// Authenticate extracts the appropriate database from a request
func Authenticate(db *database.AdminDB, r *http.Request) (database.DB, error) {
	return db, nil
}

// Request handles generation of credentials and tokens
func Request(r *http.Request) (*AuthRequest, error) {
	return nil, nil
}

// Accept accepts an auth request, and generates an associated code that can be
// exchanged for a token
func Accept(r *AuthRequest) string {
	return ""
}

// Given a request with code, generates the associated connection, and returns the api key (token)
func Token(db database.DB, r *http.Request) (string, error) {
	return "", nil
}

// Mux returns the chi mux for authentication
func Mux(db *database.AdminDB) *chi.Mux {
	return nil
}
