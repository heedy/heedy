package shell

/**

Provides the ability to list users

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"connectordb/streamdb/util/njson"
	"fmt"
)

// The clear command
type ListUsers struct {
}

func (h ListUsers) Help() string {
	return "Lists existing users"
}

func (h ListUsers) Usage() string {
	return "call with the -json flag to dump a json document of the users\n"
}

func (h ListUsers) argparse(args []string) (json bool) {
	if len(args) < 2 {
		return false
	}

	if args[1] == "-json" {
		return true
	}

	fmt.Println(Red + "Ignoring unknown argument: " + args[1] + Reset)
	return false
}

func (h ListUsers) Execute(shell *Shell, args []string) {
	json := h.argparse(args)

	users, err := shell.operator.ReadAllUsers()
	if shell.PrintError(err) {
		return
	}

	if json {
		bytes, err := njson.MarshalIndentWithTag(users, "", "  ", "")
		if shell.PrintError(err) {
			return
		}

		fmt.Printf(string(bytes))
		fmt.Println("")
	} else {
		for _, usr := range users {
			admin := "  "
			if usr.Admin {
				admin = Yellow + "* "
			}

			fmt.Printf("%s%s\t%s\t%d%s\n", admin, usr.Name, usr.Email, usr.UserId, Reset)
		}

		fmt.Print("\n\n* = admin\n")
	}
}

func (h ListUsers) Name() string {
	return "lsuser"
}
