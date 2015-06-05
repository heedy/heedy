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
type Passwd struct {
}

func (h Passwd) Help() string {
	return "Changes a user's password: 'passwd username'"
}

func (h Passwd) Usage() string {
	return ""
}

func (h Passwd) Execute(shell *Shell, args []string) {
	if len(args) < 2 {
		fmt.Println(Red + "Must supply a username" + Reset)
		return
	}

	operator := shell.operator
	username := args[1]

	fmt.Println("Enter password or blank to cancel:")
	passwd := shell.ReadRepeatPassword()
	if passwd == "" {
		return
	}

	err := operator.ChangeUserPassword(username, passwd)

	if err != nil {
		fmt.Println(Red + err.Error() + Reset)
		return
	}

	fmt.Println(Green + "Changed password for: " + args[1] + Reset)
}

func (h Passwd) Name() string {
	return "passwd"
}
