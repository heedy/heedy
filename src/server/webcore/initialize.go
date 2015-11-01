package webcore

import (
	"config"

	"github.com/gorilla/securecookie"
)

// Initialize sets up the necessary global state of WebCore such that it fits the given configuration
func Initialize(c *config.Configuration) error {
	//First initialize the sessino cookies
	authkey, err := c.GetSessionAuthKey()
	if err != nil {
		return err
	}
	encryptkey, err := c.GetSessionEncryptionKey()
	if err != nil {
		return err
	}
	CookieMonster = securecookie.New(authkey, encryptkey)

	//Now initialize the QueryTimer map
	QueryTimers = make(map[string]*QueryTimer)

	return nil
}
