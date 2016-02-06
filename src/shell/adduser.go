/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

/* Provides the ability to create users

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"fmt"
)

func init() {
	help := "An interactive session to create a new user"
	usage := "adduser"
	name := "adduser"

	main := func(shell *Shell, args []string) uint8 {
		name := shell.ReadAnswer("Enter the name for the new user: ")
		email := shell.ReadAnswer("Enter the email for the new user: ")
		role := shell.ReadAnswer("Enter the role for the new user (user): ")
		ispublic := shell.ReadAnswer("Is this user public (true): ")

		if role == "" {
			role = "user"
		}

		// Do the password check
		password := ""
		for password == "" {
			password = shell.ReadRepeatPassword()

			if password == "" {
				decision := shell.ReadAnswer("Passwords did not match, type 'yes' to try again")

				if decision != "yes" {
					return 1
				}
			}
		}

		fmt.Printf("Creating User %v at %v\n", name, email)

		err := shell.operator.CreateUser(name, email, password, role, ispublic == "true" || ispublic == "")
		if shell.PrintError(err) {
			return 1
		}

		return 0
	}

	registerShellCommand(help, usage, name, main)
}
