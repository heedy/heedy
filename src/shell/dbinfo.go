/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package shell

/* Gives information about the state of the database

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import (
	"config"
	"fmt"
)

func init() {
	help := "Prints information about the database to the console"
	usage := `Usage: dbinfo`
	name := "dbinfo"

	main := func(shell *Shell, args []string) uint8 {
		dbcxn := config.Get().Sql.GetSqlConnectionString()
		fmt.Printf("Database: %s %v\n", config.Get().Sql.Type, dbcxn)

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
