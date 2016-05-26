/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package shell

/* Grants admin to a user

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Grants admin to a user: 'grantadmin username'"
	usage := `Usage: grantadmin username`
	name := "mkadmin"

	main := func(shell *Shell, args []string) uint8 {
		if len(args) < 2 {
			fmt.Println(Red + "Must supply a name" + Reset)
			return 1
		}

		operator := shell.operator

		err := operator.UpdateUser(args[1], map[string]interface{}{"role": "admin"})
		if shell.PrintError(err) {
			return 1
		}

		fmt.Println(Green + "Granted admin to: " + args[1] + Reset)
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
