package permissions

import "errors"

// DeviceRole represents the permissions that a device with the gien role can take
// These permssions go UP TO the owning user's permissions. That is, a
// device will inherit up to its parent's permissions
type DeviceRole struct {

	// Access Levels. These are defined in the AccessLevel map of the config.
	// There are 2 levels defined by default
	// full - total permissions (everything true)
	// none - 0 permissions (everything false)

	PrivateAccessLevel string `json:"private_access_level"` // The access level to private users/devices/streams
	PublicAccessLevel  string `json:"public_access_level"`  // The access level to public users/devices/streams
	UserAccessLevel    string `json:"user_access_level"`    // The access level to devices/streams that belong to you and your own user
	SelfAccessLevel    string `json:"self_access_level"`    // The access level to give to streams that belong to querying device and to its own device
}

// Validate ensures that the given permissions have all correct values
func (r *DeviceRole) Validate(p *Permissions) error {
	if r == nil {
		return errors.New("null roles are invalid")
	}

	if _, err := p.GetAccessLevel(r.PrivateAccessLevel); err != nil {
		return err
	}
	if _, err := p.GetAccessLevel(r.PublicAccessLevel); err != nil {
		return err
	}
	if _, err := p.GetAccessLevel(r.UserAccessLevel); err != nil {
		return err
	}
	if _, err := p.GetAccessLevel(r.SelfAccessLevel); err != nil {
		return err
	}

	return nil
}
