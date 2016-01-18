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

	"github.com/gorilla/securecookie"
)

// Validate ensures that the given permissions have all correct values
func (p Permissions) Validate(c *Configuration) error {

	if _, err := c.GetAccessLevel(p.PrivateReadAccessLevel); err != nil {
		return err
	}
	if _, err := c.GetAccessLevel(p.PublicReadAccessLevel); err != nil {
		return err
	}
	if _, err := c.GetAccessLevel(p.PrivateWriteAccessLevel); err != nil {
		return err
	}
	if _, err := c.GetAccessLevel(p.PublicWriteAccessLevel); err != nil {
		return err
	}
	if _, err := c.GetAccessLevel(p.SelfWriteAccessLevel); err != nil {
		return err
	}
	if _, err := c.GetAccessLevel(p.SelfReadAccessLevel); err != nil {
		return err
	}

	return nil
}

// Validate takes a session and makes sure that all of the keys and fields are set up correctly
func (s *Session) Validate() error {
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

// Validate takes a frontend and ensures that all the necessary configuration fields are set up
// correctly.
func (f *Frontend) Validate() (err error) {

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
	if err = f.Session.Validate(); err != nil {
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

	return nil
}

// Validate takes a configuration and makes sure that it is set up correctly for use in the ConnectorDB
// database. It returns nil if the configuration is valid, and returns an error if an error was found.
// Validate also sets up any missing values to their defaults if they are not required.
func (c *Configuration) Validate() error {
	// First, make sure that the frontend is valid
	if c.Version != 1 {
		return errors.New("This version of ConnectorDB only accepts configuration version 1")
	}

	if c.BatchSize <= 0 {
		return errors.New("Batch size must be >=0")
	}
	if c.ChunkSize <= 0 {
		return errors.New("Chunk size must be >=0")
	}

	if c.UseCache {
		if c.UserCacheSize < 1 {
			return errors.New("User cache size must be >=1")
		}
		if c.DeviceCacheSize < 1 {
			return errors.New("Device cache size must be >=1")
		}
		if c.StreamCacheSize < 1 {
			return errors.New("Stream cache size must be >=1")
		}
	}

	if c.IDScramblePrime <= 0 {
		return errors.New("The ID Scramble prime must be a prime > 0.")
	}

	// Ensure that all the access level keys have valid access levels
	for key := range c.AccessLevels {
		if c.AccessLevels[key] == nil {
			return fmt.Errorf("Invalid access level '%s'", key)
		}
	}

	// Make sure the permissions are all valid
	hadNobody := false
	hadUser := false
	hadAdmin := false
	for key := range c.Permissions {
		if key == "admin" {
			hadAdmin = true
		}
		if key == "user" {
			hadUser = true
		}
		if key == "nobody" {
			hadNobody = true
		}
		if err := c.Permissions[key].Validate(c); err != nil {
			return err
		}
	}
	if !(hadNobody && hadUser && hadAdmin) {
		return errors.New("There must be at least user, admin, and nobody permissions set.")
	}

	// Now let's validate the frontend
	if err := c.Frontend.Validate(); err != nil {
		return err
	}
	return nil
}
