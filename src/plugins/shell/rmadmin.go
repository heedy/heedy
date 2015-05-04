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
type RevokeAdmin struct {
}

func (h RevokeAdmin) Help() string {
	return "Revokes admin from a user: 'rmadmin username'"
}

func (h RevokeAdmin) Usage() string {
	return h.Help()
}

func (h RevokeAdmin) Execute(shell *Shell, args []string) {
	if len(args) < 2 {
		fmt.Println(Red + "Must supply a name" + Reset)
		return
	}

	operator := shell.operator

	user, err := operator.ReadUser(args[1])
	if shell.PrintError(err) {
		return
	}

	orig := *user      // our original to revert values to
	user.Admin = false // grant admin

	err = operator.UpdateUser(user, orig)
	if shell.PrintError(err) {
		return
	}

	fmt.Println(Green + "Revoked admin from: " + args[1] + Reset)
}

func (h RevokeAdmin) Name() string {
	return "rmadmin"
}
