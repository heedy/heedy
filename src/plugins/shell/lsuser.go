package shell

/**

Provides the ability to list users

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"fmt"
)

// The clear command
type ListUsers struct {
}

func (h ListUsers) Help() string {
	return "Lists existing users"
}

func (h ListUsers) Usage() string {
	return ""
}

func (h ListUsers) Execute(shell *Shell, args []string) {

	users, err := shell.operator.ReadAllUsers()
	if shell.PrintError(err) {
		return
	}

	for _, usr := range(users) {
		admin := "  "
		if usr.Admin {
			admin = Yellow + "* "
		}

		fmt.Printf("%s%s\t%s\t%d%s\n", admin, usr.Name, usr.Email, usr.UserId, Reset)
	}

	fmt.Print("\n\n* = admin\n")
}

func (h ListUsers) Name() string {
	return "lsuser"
}
