/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package config

import (
	"errors"
	"fmt"
	"path/filepath"

	"config/permissions"

	psconfig "github.com/connectordb/pipescript/config"
)

// Validate takes a frontend and ensures that all the necessary configuration fields are set up
// correctly.
func (f *Frontend) Validate(c *Configuration) (err error) {

	// Validate the TLS config
	if err = f.TLS.Validate(); err != nil {
		return err
	}

	// Validate the Session
	if err = f.CookieSession.Validate(); err != nil {
		return err
	}

	if f.SiteURL == "" {
		f.SiteURL = f.Hostname
		// If serving to all, and site url is not set, we make the site url localhost
		// This fixes CORS issues when debugging: your browser says you're at localhost,
		// but ConnectorDB thinks it is at 0.0.0.0, so it doesn't send the correct CORS headers.
		if f.SiteURL == "" {
			f.SiteURL = "localhost"
		}
	}

	if f.InsertLimitBytes < 100 {
		return errors.New("The limit of single insert has to be at least 100 bytes.")
	}

	if err = f.Websocket.Validate(); err != nil {
		return err
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

	// Validate PipeScript
	if c.PipeScript == nil {
		c.PipeScript = psconfig.Default()
	}
	if err := c.PipeScript.Validate(); err != nil {
		return err
	}

	// Try loading the permissions
	p, err := permissions.Load(c.Permissions)
	if err != nil {
		return err
	}
	// Set the absolute path if not default
	if c.Permissions != "default" {
		c.Permissions, err = filepath.Abs(c.Permissions)
		if err != nil {
			return err
		}
	}

	// Check that the initial user permissions exist if given
	if c.InitialUser != nil && c.InitialUser.Role != "" {
		if _, ok := p.UserRoles[c.InitialUser.Role]; !ok {
			return fmt.Errorf("Could not find role of '%s' for the initial creation user", c.InitialUser.Role)
		}
	}

	if c.IDScramblePrime <= 0 {
		return errors.New("The ID Scramble prime must be a prime > 0.")
	}

	// Now see if we have a valid hashing algorithm
	if c.PasswordHash != "bcrypt" && c.PasswordHash != "SHA512" {
		return errors.New("The password hashing algorithm must be one of 'SHA512' or 'bcrypt'")
	}

	// Now let's validate the frontend
	if err := c.Frontend.Validate(c); err != nil {
		return err
	}
	return nil
}
