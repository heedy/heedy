/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package webcore

import (
	"config"

	"github.com/gorilla/securecookie"
)

// Initialize sets up the necessary global state of WebCore such that it fits the given configuration.
// Note that it is called in a config change callback, and does not use locking of state, so some weird
// bugs might be possible if config is reloaded frequently during heavy load
func Initialize(c *config.Configuration) error {
	//First initialize the sessino cookies
	authkey, err := c.Frontend.CookieSession.GetAuthKey()
	if err != nil {
		return err
	}
	encryptkey, err := c.Frontend.CookieSession.GetEncryptionKey()
	if err != nil {
		return err
	}
	CookieMonster = securecookie.New(authkey, encryptkey)

	//Set up the server globals
	AllowCrossOrigin = c.AllowCrossOrigin
	SiteName = c.GetSiteURL()

	CookieMaxAge = c.CookieSession.MaxAge

	// Set the enabled state of the server
	if c.Enabled != IsActive {
		SetEnabled(c.Enabled)
	}

	return nil
}
