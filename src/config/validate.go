/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"errors"
	"fmt"

	"config/permissions"

	psconfig "github.com/connectordb/pipescript/config"
)

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

	// Check that the initial user permissions exist if given
	if c.InitialUserRole != "" {
		if _, ok := p.UserRoles[c.InitialUserRole]; !ok {
			return fmt.Errorf("Could not find role of '%s' for the initial creation user", c.InitialUserRole)
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
