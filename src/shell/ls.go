/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

/* lists users/devices/streams

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"fmt"

	"github.com/connectordb/njson"
)

func init() {
	help := "Lists users, devices or streams given a parent path"
	usage := `Usage: ls [usr[/dev]]`
	name := "ls"

	main := func(shell *Shell, args []string) uint8 {

		path := ""
		if len(args) == 2 {
			path = shell.ResolvePath(args[1])
		}

		usr, dev, stream := shell.ReadPath(path)

		var err error
		var toPrint interface{}

		switch {
		case len(args) == 1:
			toPrint, err = shell.operator.ReadAllUsers()
		case dev == nil:
			toPrint, err = shell.operator.ReadAllDevicesByUserID(usr.UserId)
		case stream == nil:
			toPrint, err = shell.operator.ReadAllStreamsByDeviceID(dev.DeviceId)
		default:
			toPrint = []byte("You specified a full path, try cat instead.")
		}

		if shell.PrintError(err) {
			return 1
		}

		bytes, err := njson.MarshalIndentWithTag(toPrint, "", "  ", "")
		if shell.PrintError(err) {
			return 1
		}

		fmt.Printf(string(bytes))
		return 1
	}

	registerShellCommand(help, usage, name, main)
}
