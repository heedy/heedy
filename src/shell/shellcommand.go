/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

var (
	// A list of all shell commands
	allCommands []shellCommand
)

// MainFunction is sort of like the "main" you'd write in c.
// ReturnCodes are defined above
type mainFunction func(shell *Shell, args []string) uint8

// Registers a new command for all shells to use, this should be done during
// init()
func registerShellCommand(help, usage, name string, main mainFunction) {
	scw := shellCommand{help, usage, name, main}

	allCommands = append(allCommands, scw)
}

// shellCommand creates a command from a function and the associated help text.
type shellCommand struct {
	help  string
	usage string
	name  string
	main  mainFunction
}
