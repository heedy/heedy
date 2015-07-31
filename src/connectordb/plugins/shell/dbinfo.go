package shell

/**

Gives information about the state of the database

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"connectordb/config"
	"fmt"
)

// The clear command
type Dbinfo struct {
}

func (h Dbinfo) Help() string {
	return "Prints information about the database to the console"
}

func (h Dbinfo) Usage() string {
	return `Usage: dbinfo`
}

func (h Dbinfo) Execute(shell *Shell, args []string) {
	dbcxn := config.GetDatabaseConnectionString()
	fmt.Printf("Database: %v\n", dbcxn)

	users, _ := shell.operator.CountUsers()
	fmt.Printf("UserCount: %v\n", users)

	devices, _ := shell.operator.CountDevices()
	fmt.Printf("DeviceCount: %v\n", devices)

	streams, _ := shell.operator.CountStreams()
	fmt.Printf("StreamCount: %v\n", streams)
}

func (h Dbinfo) Name() string {
	return "dbinfo"
}
