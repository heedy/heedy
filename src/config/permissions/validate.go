package permissions

import (
	"errors"
	"fmt"
)

// Validate takes a permissions file and makes sure that it is set up correctly for use in the ConnectorDB
// database. It returns nil if it is valid, and returns an error if an error was found.
// Validate also sets up any missing values to their defaults if they are not required.
func (p *Permissions) Validate() error {
	if p.Version != 1 {
		return errors.New("This version of ConnectorDB only accepts permissions version 1")
	}
	// Ensure that all the access level keys have valid access levels
	for key := range p.AccessLevels {
		if p.AccessLevels[key] == nil {
			return fmt.Errorf("Invalid access level '%s'", key)
		}
	}

	// Make sure the permissions are all valid
	hadNobody := false
	for key := range p.Roles {
		if key == "nobody" {
			hadNobody = true
		}
		if err := p.Roles[key].Validate(p); err != nil {
			return err
		}
	}
	if !(hadNobody) {
		return errors.New("There must be at least nobody permissions set.")
	}

	return nil
}
