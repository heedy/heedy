/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

/* Provides the ability to create devices

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"fmt"
)

func init() {
	help := "Creates a new Device"
	usage := "adddev user/dev"
	name := "adddev"

	main := func(shell *Shell, args []string) uint8 {
		if len(args) != 2 {
			fmt.Printf(Red + "Error: Wrong number of args\n" + Reset)
			return 1
		}

		path := args[1]
		err := shell.operator.CreateDevice(path, false)

		if shell.PrintError(err) {
			return 1
		}

		fmt.Printf("Device created: %v\n", path)
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
