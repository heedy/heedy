package shell

/* Revokes admin from a user

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Revokes admin from a user: 'rmadmin username'"
	usage := `Usage: rmadmin username`
	name := "rmadmin"

	main := func(shell *Shell, args []string) uint8 {
		if len(args) < 2 {
			fmt.Println(Red + "Must supply a name" + Reset)
			return 1
		}

		err := shell.operator.SetAdmin(args[1], false)
		if shell.PrintError(err) {
			return 1
		}

		fmt.Println(Green + "Revoked admin from: " + args[1] + Reset)
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
