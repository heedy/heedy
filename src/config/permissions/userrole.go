/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package permissions

// UserRole encodes the rules that are followed for the user type
type UserRole struct {
	Join                bool   `json:"join"`                  // Whether the user can use the "join" interface to add new users (which might include captcha, etc)
	JoinDisabledMessage string `json:"join_disabled_message"` // The error message to write when join is disabled

	MaxDevices int `json:"max_devices"` // The maximum number of devices for the user. 0 is unlimited
	MaxStreams int `json:"max_streams"` // The maximum number of streams per device. 0 is unlimited

	MaxDeviceSize int64 `json:"max_device_size"` // The maximum size of a device in bytes (0 is unlimited)
	MaxStreamSize int64 `json:"max_stream_size"` // The maximum size of a stream in bytes (0 is unlimited)

	MaxPrivateDevices int `json:"max_private_devices"` // The maximum allowed private devices (0 is unlimited (since metalog is ALWAYS private))

	CanBePrivate bool `json:"user_can_be_private"` // Whether the user can be private

	// The maximum permissions that the user's devices can have
	DeviceRole
}
