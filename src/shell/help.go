/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package shell

/* Shows help info

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"fmt"
	"strings"
)

// Adds tabs to the beginning of all lines
func indent(text string) string {
	text = "\t" + text
	text = strings.Replace(text, "\n", "\n\t", -1)
	return text
}

func init() {
	help := "Shows all commands or help about a specific one."
	usage := `help [commandname]

The optional command name will show more detailed information about a given
command.
`
	name := "help"

	main := func(shell *Shell, args []string) uint8 {
		if len(args) == 2 {
			for _, cmd := range allCommands {
				if cmd.name == args[1] {
					fmt.Println()
					shell.PrintTitle("NAME")
					fmt.Printf("\t%v - %v\n\n", cmd.name, cmd.help)

					shell.PrintTitle("USAGE")
					fmt.Println(indent(cmd.usage))

					return 0
				}
			}
			shell.PrintErrorText("%s not found, listing known commands:", args[1])
		}

		shell.PrintTitle("ConnectorDB Shell Help")

		for _, cmd := range allCommands {
			fmt.Printf("%v\t- %v\n", cmd.name, cmd.help)
		}
		fmt.Println("")
		fmt.Println("Use 'help [commandname]' to show help for a specific command.")
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
