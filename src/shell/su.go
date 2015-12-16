/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

/* Starts a new shell as a given user to contain permissions

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"connectordb/operator"
	"fmt"
)

func init() {
	help := "Changes priviliges to a particular user: 'su username'"
	usage := `Usage: su username`
	name := "su"

	main := func(shell *Shell, args []string) uint8 {
		if len(args) < 2 {
			fmt.Println(Red + "Must supply a name" + Reset)
			return 1
		}

		username := args[1]

		suOperator, err := operator.NewUserOperator(shell.sdb, username)
		if shell.PrintError(err) {
			return 1
		}

		sushell := CreateShell(shell.sdb)
		sushell.operator = suOperator
		sushell.operatorName = username

		sushell.Repl()
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
