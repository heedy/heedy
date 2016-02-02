package permissions

var (
	// FullAccessLevel is the access level of a total database admin. EVERYTHING is accessible
	FullAccessLevel = AccessLevel{true, true, true, true, true, true, true, true, true, "full", "full"}
	// NoneAccessLevel is the access level of a total nobody - can't access jack shit
	NoneAccessLevel = AccessLevel{
		ReadAccess:  "none",
		WriteAccess: "none",
	}
)

// AccessLevel represents the total access that a certain object has to the ConnectorDB database
type AccessLevel struct {
	CanCreateUser   bool `json:"can_create_user"`
	CanCreateDevice bool `json:"can_create_device"`
	CanCreateStream bool `json:"can_create_stream"`

	CanDeleteUser   bool `json:"can_delete_user"`
	CanDeleteDevice bool `json:"can_delete_device"`
	CanDeleteStream bool `json:"can_delete_stream"`

	CanListUsers   bool `json:"can_list_users"`
	CanListDevices bool `json:"can_list_devices"`
	CanListStreams bool `json:"can_list_streams"`

	// ReadAccess and WriteAccess are strings which are names of RWAccess objects
	// defined in rw_access
	ReadAccess  string `json:"read_access"`
	WriteAccess string `json:"write_access"`
}

// Validate ensures that the given access level has all of its setting valid
func (a *AccessLevel) Validate(p *Permissions) error {
	if _, err := p.GetRWAccess(a.ReadAccess); err != nil {
		return err
	}
	if _, err := p.GetRWAccess(a.WriteAccess); err != nil {
		return err
	}
	return nil
}
