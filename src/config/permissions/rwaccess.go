/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package permissions

import "reflect"

var (
	// NoneRWAccess is the RW permission to give a device when it does not have ANY permissions associated with an action
	NoneRWAccess = RWAccess{}
	// FullRWAccess is the RW permission to give a total administrator - everything is accessible
	FullRWAccess = RWAccess{true, true, true, true,
		true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true, true,
		true, true, true, true, true, true, true, true, true, true, nil}
)

// RWAccess is a struct of boolean permissions given for a certain role.
type RWAccess struct {

	// General access level options
	CanAccessUser   bool `json:"can_access_user"`
	CanAccessDevice bool `json:"can_access_device"`
	CanAccessStream bool `json:"can_access_stream"`
	// Read/write of streams
	CanAccessStreamData bool `json:"can_access_stream_data"`

	// Whether or not this is allowed to write non-user-editable devices
	// For use in admin
	CanAccessNonUserEditableDevices bool `json:"can_access_non_user_editable_devices"`

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
	DeviceName         bool `json:"device_name"`
	DeviceNickname     bool `json:"device_nickname"`
	DeviceDescription  bool `json:"device_description"`
	DeviceIcon         bool `json:"device_icon"`
	DeviceAPIKey       bool `json:"device_apikey"`
	DeviceEnabled      bool `json:"device_enabled"`
	DeviceIsVisible    bool `json:"device_visible"`
	DeviceUserEditable bool `json:"device_user_editable"`
	DevicePublic       bool `json:"device_public"`
	DeviceRole         bool `json:"device_role"`

	// Access of stream properties
	StreamName        bool `json:"stream_name"`
	StreamNickname    bool `json:"stream_nickname"`
	StreamDescription bool `json:"stream_description"`
	StreamIcon        bool `json:"stream_icon"`
	StreamSchema      bool `json:"stream_schema"`
	StreamDatatype    bool `json:"stream_datatype"`
	StreamEphemeral   bool `json:"stream_ephemeral"`
	StreamDownlink    bool `json:"stream_downlink"`

	// Internal: cached map of access levels (used in reflection)
	cmap map[string]bool
}

// LoadMap generates a map of the access levels by their json attribute names.
// Note that access levels are entirely boolean
func (a *RWAccess) LoadMap() error {
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
func (a *RWAccess) GetMap() map[string]bool {
	if a.cmap == nil {
		a.LoadMap()
	}
	return a.cmap
}

func (a *RWAccess) Validate() error {
	return a.LoadMap()
}
