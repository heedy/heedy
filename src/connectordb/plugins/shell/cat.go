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
type Cat struct {
}

func (h Cat) Help() string {
	return "Prints information about a user/stream/device to the console"
}

func (h Cat) Usage() string {
	return `Usage: cat username

	Prints information about the user to the standard output in a JSON format.
	`
}

func (h Cat) Execute(shell *Shell, args []string) {
	if len(args) < 2 {
		fmt.Println(Red + "Must supply a name" + Reset)
		return
	}

	operator := shell.operator
	user, err := operator.ReadUser(args[1])
	if shell.PrintError(err) {
		return
	}

	bytes, err := njson.MarshalIndentWithTag(user, "", "  ", "DOESNOTEXIST")
	if shell.PrintError(err) {
		return
	}

	fmt.Printf(string(bytes))
	fmt.Println("")
}

func (h Cat) Name() string {
	return "cat"
}
