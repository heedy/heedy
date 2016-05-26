/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package shell

/* Clears the screen

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Clears the screen"
	usage := `Usage: clear`
	name := "clear"

	main := func(shell *Shell, args []string) uint8 {
		fmt.Println(Reset)
		shell.Cls()
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
