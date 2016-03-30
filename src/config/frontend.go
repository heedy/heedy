/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"fmt"
	"strings"
)

// Captcha allows using reCaptcha to ensure logins are real users
type Captcha struct {
	Enabled    bool   `json:"enabled"`
	SiteKey    string `json:"site_key"`
	SiteSecret string `json:"site_secret"`
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
	SiteURL string `json:"siteurl"`

	// Whether the site options permit CORS
	AllowCrossOrigin bool `json:"allowcrossorigin"`

	// The session cookies to allow in the website
	CookieSession CookieSession `json:"cookie"`

	// This enables TLS on the server
	TLS TLS `json:"tls"`

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
	InsertLimitBytes int64 `json:"insert_limit_bytes"`

	// Options for websocket connections
	Websocket Websocket `json:"websocket"`

	// Minify gives us whether ConnectorDB should minify the templates that are run.
	// At this point, only the templates hav minify support - static files are not minifed
	Minify bool `json:"minify"`
}

// TLSEnabled returns whether or not TLS os enabled for the frontend
func (f *Frontend) TLSEnabled() bool {
	return f.TLS.Enabled
}

// GetSiteURL returns a URL to the frontend
func (f *Frontend) GetSiteURL() string {
	siteurl := f.SiteURL

	if !strings.HasPrefix(siteurl, "http") {
		// If the domain given starts with http, we assume the full correct answer was
		// set - and we don't worry about setting up a good url.
		// Otherwise, set up the URL according to the current port setup
		siteurl = "http"

		if f.TLSEnabled() {
			siteurl += "s"
		}
		siteurl += "://" + f.SiteURL
		if !(f.TLSEnabled() && f.Port == 443) || (!f.TLSEnabled() && f.Port == 80) {
			// If it is NOT a standard port, then add the port number to the URL
			siteurl = fmt.Sprintf("%s:%d", siteurl, f.Port)
		}
	}

	// If it ends with a slash, remove the slash
	if strings.HasSuffix(siteurl, "/") {
		siteurl = siteurl[0 : len(siteurl)-2]
	}

	return siteurl
}
