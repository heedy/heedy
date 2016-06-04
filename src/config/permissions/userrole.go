/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package permissions

// BaseDefaults represents the fields that are present in users, devices and streams
type BaseDefaults struct {
	Nickname    string `json:"nickname"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// UserDefaults represent default values for users
type UserDefaults struct {
	BaseDefaults
	Role   string `json:"role"`
	Public bool   `json:"public"`
}

// DeviceDefaults is the default
type DeviceDefaults struct {
	BaseDefaults
	Role         string `json:"role"`
	Enabled      bool   `json:"enabled"`
	Public       bool   `json:"public"`
	IsVisible    bool   `json:"visible"`
	UserEditable bool   `json:"user_editable"`
}

type StreamDefaults struct {
	BaseDefaults
	Schema    string `json:"schema"`
	Datatype  string `json:"datatype"`
	Ephemeral bool   `json:"ephemeral"`
	Downlink  bool   `json:"downlink"`
}

// UserRole encodes the rules that are followed for the user type
type UserRole struct {
	Join                bool   `json:"join"`                  // Whether the user can use the "join" interface to add new users (which might include captcha, etc)
	JoinRole            string `json:"join_role"`             // The role to use for users joining
	JoinDisabledMessage string `json:"join_disabled_message"` // The error message to write when join is disabled

	MaxDevices int64 `json:"max_devices"` // The maximum number of devices for the user. 0 is unlimited
	MaxStreams int64 `json:"max_streams"` // The maximum number of streams per device. 0 is unlimited

	MaxDeviceSize int64 `json:"max_device_size"` // The maximum size of a device in bytes (0 is unlimited)
	MaxStreamSize int64 `json:"max_stream_size"` // The maximum size of a stream in bytes (0 is unlimited)

	MaxPrivateDevices int64 `json:"max_private_devices"` // The maximum allowed private devices (0 is unlimited (since metalog is ALWAYS private))

	CanBePrivate bool `json:"user_can_be_private"` // Whether the user can be private

	CreateUserDefaults   UserDefaults   `json:"create_user_defaults"`
	CreateDeviceDefaults DeviceDefaults `json:"create_device_defaults"`
	CreateStreamDefaults StreamDefaults `json:"create_stream_defaults"`

	// The maximum permissions that the user's devices can have
	DeviceRole
}
