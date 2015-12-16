/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package shell

/* Gives information about the state of the database

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Prints information about the database to the console"
	usage := `Usage: dbinfo`
	name := "dbinfo"

	main := func(shell *Shell, args []string) uint8 {
		dbcxn := cfg.GetSqlConnectionString()
		fmt.Printf("Database: %v\n", dbcxn)

		users, _ := shell.operator.CountUsers()
		fmt.Printf("UserCount: %v\n", users)

		devices, _ := shell.operator.CountDevices()
		fmt.Printf("DeviceCount: %v\n", devices)

		streams, _ := shell.operator.CountStreams()
		fmt.Printf("StreamCount: %v\n", streams)
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
