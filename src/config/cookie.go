package config

import (
	"encoding/base64"
	"errors"

	"github.com/gorilla/securecookie"
)

// CookieSession refers to a cookie session
type CookieSession struct {
	AuthKey       string `json:"authkey"`       //The key used to sign sessions
	EncryptionKey string `json:"encryptionkey"` //The key used to encrypt sessions in cookies
	MaxAge        int    `json:"maxage"`        //The maximum age of a cookie in a session (seconds)
}

// GetAuthKey returns the bytes associated with the config string
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

// GetEncryptionKey returns the bytes associated with the config string
func (s *CookieSession) GetEncryptionKey() ([]byte, error) {
	//If no session encryption key is in config, generate one
	if s.EncryptionKey == "" {
		return securecookie.GenerateRandomKey(32), nil
	}

	return base64.StdEncoding.DecodeString(s.EncryptionKey)
}
