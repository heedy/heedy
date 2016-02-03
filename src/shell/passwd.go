/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

/* Allows us to reset a user's password

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Changes a user's password 'passwd username'"
	usage := `Usage: passwd username`
	name := "passwd"

	main := func(shell *Shell, args []string) uint8 {
		if len(args) < 2 {
			fmt.Println(Red + "Must supply a username" + Reset)
			return 1
		}

		operator := shell.operator
		username := args[1]

		fmt.Println("Enter password or blank to cancel:")
		passwd := shell.ReadRepeatPassword()
		if passwd == "" {
			return 1
		}

		err := operator.UpdateUser(username, map[string]interface{}{"password": passwd})

		if shell.PrintError(err) {
			return 0
		}

		fmt.Println(Green + "Changed password for: " + args[1] + Reset)
		return 1
	}

	registerShellCommand(help, usage, name, main)
}
