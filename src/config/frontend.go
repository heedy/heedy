/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

// Validate takes a session and makes sure that all of the keys and fields are set up correctly
func (s *CookieSession) Validate() error {
	if s.AuthKey == "" {
		sessionAuthkey := securecookie.GenerateRandomKey(64)
		s.AuthKey = base64.StdEncoding.EncodeToString(sessionAuthkey)
	}
	if s.EncryptionKey == "" {
		sessionEncKey := securecookie.GenerateRandomKey(32)
		s.EncryptionKey = base64.StdEncoding.EncodeToString(sessionEncKey)
	}

	if s.MaxAge < 0 {
		return errors.New("Max Age for cookie must be >=0")
	}

	return nil
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

	// The limit in bytes per REST insert
	InsertLimitBytes           int64 `json:"insert_limit_bytes"`
	WebsocketMessageLimitBytes int64 `json:"websocket_message_limit_bytes"`

	// The time to wait on a socket write
	WebsocketWriteWait time.Duration `json:"websocket_write_wait"`

	// Websockets ping each other to keep the connection alive
	// This sets the number od seconds between pings
	WebsocketPongWait   time.Duration `json:"websocket_pong_wait"`
	WebsocketPingPeriod time.Duration `json:"websocket_ping_period"`

	// The websocket read/write buffer for socket upgrader
	WebsocketReadBufferSize  int `json:"websocket_read_buffer"`
	WebsocketWriteBufferSize int `json:"websocket_write_buffer"`

	// The number of messages to buffer
	WebsocketMessageBuffer int64 `json:"websocket_message_buffer"`

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

// Validate takes a frontend and ensures that all the necessary configuration fields are set up
// correctly.
func (f *Frontend) Validate(c *Configuration) (err error) {

	if f.TLSEnabled() {
		// If both key and cert are given, assume that we want to use TLS
		_, err = tls.LoadX509KeyPair(f.TLSCert, f.TLSKey)
		if err != nil {
			return err
		}

		//Set the file paths to be full paths
		f.TLSCert, err = filepath.Abs(f.TLSCert)
		if err != nil {
			return err
		}
		f.TLSKey, err = filepath.Abs(f.TLSKey)
		if err != nil {
			return err
		}
	}

	// Validate the Session
	if err = f.CookieSession.Validate(); err != nil {
		return err
	}

	// Set up the optional configuration parameters

	if f.Hostname == "" {
		f.Hostname, err = os.Hostname()
		if err != nil {
			f.Hostname = "localhost"
		}
	}

	if f.Domain == "" {
		f.Domain = f.Hostname
	}

	if f.InsertLimitBytes < 100 {
		return errors.New("The limit of single insert has to be at least 100 bytes.")
	}

	if f.WebsocketMessageLimitBytes < 100 {
		return errors.New("The limit of a websocket message has to be at least 100 bytes.")
	}

	if f.WebsocketWriteWait < 1 {
		return errors.New("The websocket write wait time must be at least 1 second")
	}

	if f.WebsocketPongWait < 1 {
		return errors.New("The pong wait time for websocket must be at least 1s.")
	}

	if f.WebsocketPingPeriod < 1 {
		return errors.New("Websocket ping period must be at least 1 second")
	}

	if f.WebsocketMessageBuffer < 1 {
		return errors.New("The websocket message buffer must have at least one message")
	}

	if f.WebsocketWriteBufferSize < 10 {
		return errors.New("Websocket write buffer must be at least 10 bytes")
	}

	if f.WebsocketReadBufferSize < 10 {
		return errors.New("The websocket read buffer must be at least 10 bytes")
	}

	return nil
}
