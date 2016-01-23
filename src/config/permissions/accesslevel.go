/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package permissions

import "reflect"

var (
	// The NoneAccessLevel is the permission to give a device when it does not have ANY permissions associated with an action
	NoneAccessLevel = AccessLevel{}
	// The FullAccessLevel is the permission to give a total administrator - everything is accessible
	FullAccessLevel = AccessLevel{true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true, true, true, true, true, true, nil}
)

// AccessLevel is a struct of boolean permissions given for a certain role.
type AccessLevel struct {

	// General access level options
	CanAccessUser   bool `json:"can_access_user"`
	CanAccessDevice bool `json:"can_access_device"`
	CanAccessStream bool `json:"can_access_stream"`

	// Whether or not this is allowed to write non-user-editable devices
	// For use in admin
	CanAccessNonUserEditableDevices bool `json:"can_access_non_user_editable_devices"`

	// Read/write of streams
	CanReadStreamData  bool `json:"can_read_stream_data"`
	CanWriteStreamData bool `json:"can_write_stream_data"`

	// Access of user properties
	UserName        bool `json:"user_name"`
	UserNickname    bool `json:"user_nickname"`
	UserEmail       bool `json:"user_email"`
	UserDescription bool `json:"user_description"`
	UserIcon        bool `json:"user_icon"`
	UserRole        bool `json:"user_role"`
	UserPublic      bool `json:"user_public"`
	UserPassword    bool `json:"user_password"`

	// Access of device properties
	DeviceName                    bool `json:"device_name"`
	DeviceNickname                bool `json:"device_nickname"`
	DeviceDescription             bool `json:"device_description"`
	DeviceIcon                    bool `json:"device_icon"`
	DeviceAPIKey                  bool `json:"device_apikey"`
	DeviceEnabled                 bool `json:"device_enabled"`
	DeviceIsVisible               bool `json:"device_visible"`
	DeviceUserEditable            bool `json:"device_user_editable"`
	DevicePublic                  bool `json:"device_public"`
	DeviceCanReadUser             bool `json:"device_can_read_user"`
	DeviceCanReadExternal         bool `json:"device_can_read_external"`
	DeviceCanWriteUser            bool `json:"device_can_write_user"`
	DeviceCanWriteExternal        bool `json:"device_can_write_external"`
	DeviceCanReadUserStreams      bool `json:"device_can_read_user_streams"`
	DeviceCanReadExternalStreams  bool `json:"device_can_read_external_streams"`
	DeviceCanWriteUserStreams     bool `json:"device_can_write_user_streams"`
	DeviceCanWriteExternalStreams bool `json:"device_can_write_external_streams"`

	// Access of stream properties
	StreamName        bool `json:"stream_name"`
	StreamNickname    bool `json:"stream_nickname"`
	StreamDescription bool `json:"stream_description"`
	StreamIcon        bool `json:"stream_icon"`
	StreamSchema      bool `json:"stream_schema"`
	StreamEphemeral   bool `json:"stream_ephemeral"`
	StreamDownlink    bool `json:"stream_downlink"`

	// Internal: cached map of access levels (used in reflection)
	cmap map[string]bool
}

// LoadMap generates a map of the access levels by their json attribute names.
// Note that access levels are entirely boolean
func (a *AccessLevel) LoadMap() error {
	cmap := make(map[string]bool)
	atype := reflect.TypeOf(*a)
	aval := reflect.ValueOf(*a)

	for i := 0; i < aval.NumField(); i++ {
		t := atype.Field(i)
		v := aval.Field(i)
		if v.Kind() == reflect.Bool {
			cmap[t.Tag.Get("json")] = v.Bool()
		}
	}
	a.cmap = cmap
	return nil
}

// GetMap returns the map of json-values with their boolean
func (a *AccessLevel) GetMap() map[string]bool {
	if a.cmap == nil {
		a.LoadMap()
	}
	return a.cmap
}

func (a *AccessLevel) Validate() error {
	return a.LoadMap()
}
