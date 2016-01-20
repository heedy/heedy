/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"errors"
	"fmt"

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

	// Check that the initial user permissions exist if given
	if c.InitialUserPermissions != "" {
		if _, ok := c.Permissions[c.InitialUserPermissions]; !ok {
			return fmt.Errorf("Could not find permissions of '%s' for the initial creation user", c.InitialUserPermissions)
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
	for key := range c.Permissions {
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
	if !(hadNobody && hadUser) {
		return errors.New("There must be at least user and nobody permissions set.")
	}

	// Now let's validate the frontend
	if err := c.Frontend.Validate(c); err != nil {
		return err
	}
	return nil
}
