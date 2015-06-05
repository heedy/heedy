package shell

/**

Provides the ability to list users/devices/streams

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"connectordb/streamdb/util/njson"
	"fmt"
)

// The clear command
type Ls struct {
}

func (h Ls) Help() string {
	return "Lists information about a user/device/stream: 'ls path'"
}

func (h Ls) Usage() string {
	return "ls [-json] usr[/dev[/stream]]"
}

func (h Ls) Execute(shell *Shell, args []string) {

	path := ""
	if len(args) == 2 {
		path = shell.ResolvePath(args[1])
	}
	json := false
	if len(args) == 3 {
		if args[1] == "-json" {
			json = true
			path = shell.ResolvePath(args[2])
		} else {
			fmt.Println(Red + "unknown command: " + args[1] + Reset)
			return
		}
	}

	usr, dev, stream := shell.ReadPath(path)

	var err error

	switch {
	case usr == nil:
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

	case dev == nil:
		devs, err := shell.operator.ReadAllDevicesByUserID(usr.UserId)
		if shell.PrintError(err) {
			return
		}

		bytes, err := njson.MarshalIndentWithTag(devs, "", "  ", "")
		if shell.PrintError(err) {
			return
		}

		fmt.Printf(string(bytes))
		fmt.Println("")

	case stream == nil:
		streams, err := shell.operator.ReadAllStreamsByDeviceID(dev.DeviceId)
		if shell.PrintError(err) {
			return
		}

		bytes, err := njson.MarshalIndentWithTag(streams, "", "  ", "")
		if shell.PrintError(err) {
			return
		}

		fmt.Printf(string(bytes))
		fmt.Println("")
	}

	if err != nil {
		fmt.Println(Red + err.Error() + Reset)
		return
	}

}

func (h Ls) Name() string {
	return "ls"
}
