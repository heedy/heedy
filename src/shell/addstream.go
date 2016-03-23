/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

/* Provides the ability to create streams

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"connectordb/users"
	"fmt"
)

func init() {
	help := "Creates a new Stream at the given path, default is numerical."
	usage := "addstream path [type]"
	name := "addstream"

	main := func(shell *Shell, args []string) uint8 {
		path := ""
		streamType := `{"type":"number"}`

		switch len(args) {
		default:
			fmt.Printf(Red + "Error: Wrong number of args\n" + Reset)
			return 1
		case 2:
			path = args[1]
		case 3:
			path = args[1]
			streamType = args[2]
		}

		path = shell.ResolvePath(path)

		fmt.Printf("Creating Stream %v\n", path)
		err := shell.operator.CreateStream(path, &users.StreamMaker{Stream: users.Stream{Schema: streamType}})

		if shell.PrintError(err) {
			return 2
		}

		return 0
	}

	registerShellCommand(help, usage, name, main)
}
