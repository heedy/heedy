/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

// Permissions are the rules that are followed for the user type
type Permissions struct {
	Join                bool   `json:"join"`                  // Whether the user can use the "join" interface to add new users (which might include captcha, etc)
	JoinDisabledMessage string `json:"join_disabled_message"` // The error message to write when join is disabled

	MaxDevices int `json:"max_devices"` // The maximum number of devices for the user. 0 is unlimited
	MaxStreams int `json:"max_streams"` // The maximum number of streams per device. 0 is unlimited

	MaxDeviceSize int64 `json:"max_device_size"` // The maximum size of a device in bytes (0 is unlimited)
	MaxStreamSize int64 `json:"max_stream_size"` // The maximum size of a stream in bytes (0 is unlimited)

	MaxPrivateDevices int `json:"max_private_devices"` // The maximum allowed private devices (0 is unlimited (since metalog is ALWAYS private))

	CanBePrivate  bool `json:"user_can_be_private"`   // Whether the user can be private
	CreatePrivate bool `json:"user_creation_private"` // Whether the user is created as a private user

	// Access Levels. These are defined in the AccessLevel map of the config.
	// There are 2 levels defined by default
	// full - total permissions (everything true)
	// none - 0 permissions (everything false)

	PublicReadAccessLevel  string `json:"public_read_access_level"`  // The access level to public users/devices/streams
	PrivateReadAccessLevel string `json:"private_read_access_level"` // The access level to private users/devices/streams
	SelfReadAccessLevel    string `json:"self_read_access_level"`    // The access level to read self. Note that with some permissions, might be able to change self type.

	PublicWriteAccessLevel  string `json:"public_write_access_level"`  // The access level to public users/devices/streams
	PrivateWriteAccessLevel string `json:"private_write_access_level"` // The access level to private users/devices/streams
	SelfWriteAccessLevel    string `json:"self_write_access_level"`    // The access level to read self. Note that with some permissions, might be able to change self type.
}
