/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"encoding/base64"
	"fmt"

	"github.com/gorilla/securecookie"
)

// Captcha allows using reCaptcha to ensure logins are real users
type Captcha struct {
	Enabled    bool   `json:"enabled"`
	SiteKey    string `json:"site_key"`
	SiteSecret string `json:"site_secret"`
}

// CookieSession refers to a cookie session
type CookieSession struct {
	AuthKey       string `json:"authkey"`       //The key used to sign sessions
	EncryptionKey string `json:"encryptionkey"` //The key used to encrypt sessions in cookies
	MaxAge        int    `json:"maxage"`        //The maximum age of a cookie in a session (seconds)
}

// GetSessionAuthKey returns the bytes associated with the config string
func (s *CookieSession) GetAuthKey() ([]byte, error) {
	//If no session key is in config, generate one
	if s.AuthKey == "" {
		return securecookie.GenerateRandomKey(64), nil
	}

	return base64.StdEncoding.DecodeString(s.AuthKey)
}

// GetSessionEncryptionKey returns the bytes associated with the config string
func (s *CookieSession) GetEncryptionKey() ([]byte, error) {
	//If no session encryption key is in config, generate one
	if s.EncryptionKey == "" {
		return securecookie.GenerateRandomKey(32), nil
	}

	return base64.StdEncoding.DecodeString(s.EncryptionKey)
}

// Frontend represents the ConnectorDB frontend server options
type Frontend struct {

	// The hostname and port to run ConnectorDB on
	Hostname string `json:"hostname"`
	Port     uint16 `json:"port"`

	// Whether or not the frontend is enabled
	Enabled bool `json:"frontend_enabled"`

	// The domain name of the website at which connectordb is running.
	// This enables Connectordb to be able to output links to itself.
	// Leave blank if domain is the same as Hostname
	Domain string `json:"sitename"`

	// Whether the site options permit CORS
	AllowCrossOrigin bool `json:"allowcrossorigin"`

	// The session cookies to allow in the website
	CookieSession CookieSession `json:"cookie"`

	// These two options enable https on the server. Both files must exist
	// for TLS to be enabled
	TLSKey  string `json:"tls_key"`
	TLSCert string `json:"tls_cert"`

	Captcha Captcha `json:"captcha"`

	// The QueryDisplayTimer is how often to display aggregate query numbers (is seconds) in the log
	// This is a simple one-line summary of how many requests were processed.
	// Note that the change will not come into effect immediately if modified during runtime, there will be a delay before
	// the change catches on
	QueryDisplayTimer int64 `json:"query_display_timer"`
	// StatsDisplayTimer is how often to display server query statistics (in seconds). These are detailed
	// timing information for all queries, including how long they take and their standard deviations.
	// Changing during run time does not come into effect immediately: there is a delay before the change catches on.
	StatsDisplayTimer int64 `json:"stats_display_timer"`

	// Minify gives us whether ConnectorDB should minify the templates that are run.
	// At this point, only the templates hav minify support - static files are not minifed
	Minify bool `json:"minify"`
}

// TLSEnabled returns whether or not TLS os enabled for the frontend
func (f *Frontend) TLSEnabled() bool {
	return f.TLSCert != "" && f.TLSKey != ""
}

// SiteURL returns a URL to the frontend
func (f *Frontend) SiteURL() string {
	siteurl := "http"

	if f.TLSEnabled() {
		siteurl += "s"
	}
	siteurl += "://" + f.Domain

	if !(f.TLSEnabled() && f.Port == 443) || (!f.TLSEnabled() && f.Port == 80) {
		// If it is NOT a standard port, then add the port number to the URL
		siteurl = fmt.Sprintf("%s:%d", siteurl, f.Port)
	}
	return siteurl
}
