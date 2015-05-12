package shell

/**

Provides the ability to list devices for a user

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"fmt"
	"streamdb/util/njson"
)

// The clear command
type ListDevices struct {
}

func (h ListDevices) Help() string {
	return "Lists a user's devices"
}

func (h ListDevices) Usage() string {
	return `Lists devices for a given user:

	lsdev [username]

You can also output devices in JSON format:

	lsdev -json [username]

`
}

func (h ListDevices) Execute(shell *Shell, args []string) {

	if len(args) < 2 {
		fmt.Println(Red + "Must supply a name" + Reset)
		return
	}

	name := args[1]
	json := false
	if len(args) == 3 {
		if args[1] == "-json" {
			json = true
			name = args[2]
		} else {
			fmt.Println(Red + "unknown command: " + args[1] + Reset)
			return
		}
	}

	devices, err := shell.operator.ReadAllDevices(name)
	if shell.PrintError(err) {
		return
	}

	if json {
		bytes, err := njson.MarshalIndentWithTag(devices, "", "  ", "DOESNOTEXIST")
		if shell.PrintError(err) {
			return
		}

		fmt.Printf(string(bytes))
		fmt.Println("")
	} else {
		fmt.Printf("  Id\tName\tApiKey\t\n")
		for _, dev := range devices {
			admin := "  "
			if dev.IsAdmin {
				admin = Yellow + "* "
			}

			visible := ""
			if !dev.IsVisible {
				visible = Cyan + "(invisible)"
			}

			fmt.Printf("%s%d\t%s\t%s\t%s\n"+Reset, admin, dev.DeviceId, dev.Name, dev.ApiKey, visible)
		}

		fmt.Print("\n\n* = admin\n")
	}
}

func (h ListDevices) Name() string {
	return "lsdev"
}
