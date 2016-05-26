/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package shell

/* Quits the shell

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Quits the running interactive session."
	usage := `Usage: exit`
	name := "exit"

	main := func(shell *Shell, args []string) uint8 {
		fmt.Printf("exit\n")
		shell.running = false
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
