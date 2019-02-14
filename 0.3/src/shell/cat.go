/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package shell

/* Provides the ability to list users/devices/streams in JSON

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"fmt"

	"github.com/connectordb/njson"
)

func init() {
	help := "Prints information about a user/stream/device to the console"
	usage := `Usage: cat path

	Prints information about the user to the standard output in a JSON format.
	`
	name := "cat"

	main := func(shell *Shell, args []string) uint8 {
		var err error
		var toPrint interface{}

		if len(args) < 2 {
			fmt.Println(Red + "Must supply a path" + Reset)
			return 1
		}

		path := shell.ResolvePath(args[1])
		usr, dev, stream := shell.ReadPath(path)

		switch {
		default:
			toPrint = ""
		case stream != nil:
			toPrint = stream
		case dev != nil:
			toPrint = dev
		case usr != nil:
			toPrint = usr
		}

		bytes, err := njson.MarshalIndentWithTag(toPrint, "", "  ", "")
		if shell.PrintError(err) {
			return 1
		}

		fmt.Println(string(bytes))
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
