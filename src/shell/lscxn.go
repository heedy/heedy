package shell

/* Lists connection info for redis, gnats and sql

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Lists the connection addresses to the components of the system."
	usage := `Usage: lscxn`
	name := "lscxn"

	main := func(shell *Shell, args []string) uint8 {
		dbcxn := cfg.GetSqlConnectionString()
		fmt.Printf("Database: %v\n", dbcxn)

		streamdb := cfg.DatabaseDirectory
		fmt.Printf("Streamdb: %v\n", streamdb)

		redis := cfg.GetRedisURI()
		fmt.Printf("Redis: %v\n", redis)

		gnatsd := cfg.GetGnatsdURI()
		fmt.Printf("Gnatsd: %v\n", gnatsd)
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
