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
	return "ls [usr[/dev[/stream]]]"
}

func (h Ls) Execute(shell *Shell, args []string) {

	path := ""
	if len(args) == 2 {
		path = shell.ResolvePath(args[1])
	}

	usr, dev, stream := shell.ReadPath(path)

	var err error
	var toPrint interface{}

	switch {
	case len(args) == 1:
		toPrint, err = shell.operator.ReadAllUsers()
	case dev == nil:
		toPrint, err = shell.operator.ReadAllDevicesByUserID(usr.UserId)
	case stream == nil:
		toPrint, err = shell.operator.ReadAllStreamsByDeviceID(dev.DeviceId)
	default:
		toPrint = []byte("You specified a full path, try cat instead.")
	}

	if shell.PrintError(err) {
		return
	}

	bytes, err := njson.MarshalIndentWithTag(toPrint, "", "  ", "")
	if shell.PrintError(err) {
		return
	}

	fmt.Printf(string(bytes))
	fmt.Println("")
}

func (h Ls) Name() string {
	return "ls"
}
