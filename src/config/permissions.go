/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

// Permissions are the rules that are followed for the user type
type Permissions struct {
	Join bool `json:"join"` // Whether the user can use the "join" interface to add new users (which might include captcha, etc)

	//MaxDevices int // The maximum number of devices for the user
	//MaxStreams int // The maximum number of streams per device
}
