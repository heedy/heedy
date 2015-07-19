package shell

/**

Provides the ability to list users

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"fmt"

	"github.com/connectordb/njson"
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
	var err error
	var toPrint interface{}

	if len(args) < 2 {
		fmt.Println(Red + "Must supply a path" + Reset)
		return
	}

	path := shell.ResolvePath(args[1])
	usr, dev, stream := shell.ReadPath(path)

	switch {
	default:
		toPrint = ""
	case stream != nil:
		toPrint = stream
	case dev != nil:
		toPrint = dev
	case usr != nil:
		toPrint = usr
	}

	bytes, err := njson.MarshalIndentWithTag(toPrint, "", "  ", "")
	if shell.PrintError(err) {
		return
	}

	fmt.Println(string(bytes))
}

func (h Cat) Name() string {
	return "cat"
}
