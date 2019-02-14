/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package util

import (
	"bytes"
	"net"
)

const (
	// Localhost variables, both net and string
	LocalhostIpV4 = "127.0.0.1"
	LocalhostIpV6 = "0:0:0:0:0:0:0:1"
)

// Checks if the given address is for the localhost, returns false on
// lookup error
func IsLocalhost(hostname string) bool {
	ips, err := net.LookupIP(hostname)

	if err != nil {
		return false
	}

	localV4 := net.ParseIP(LocalhostIpV4)
	localV6 := net.ParseIP(LocalhostIpV6)

	for _, ip := range ips {
		if bytes.Compare(ip, localV4) == 0 || bytes.Compare(ip, localV6) == 0 {
			return true
		}
	}

	return false
}
