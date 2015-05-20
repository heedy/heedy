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
type Su struct {
}

func (h Su) Help() string {
	return "Changes priviliges to a particular user: 'su username'"
}

func (h Su) Usage() string {
	return ""
}

func (h Su) Execute(shell *Shell, args []string) {
	if len(args) < 2 {
		fmt.Println(Red + "Must supply a name" + Reset)
		return
	}

	username := args[1]

	suOperator, err := shell.sdb.GetOperator(username)
	if shell.PrintError(err) {
		return
	}

	sushell := CreateShell(shell.sdb)
	sushell.operator = suOperator
	sushell.operatorName = username

	sushell.Repl()
}

func (h Su) Name() string {
	return "su"
}
