package shell

import (
	"connectordb/config"
	"fmt"
)

// The Exit command
type LsCxn struct {
}

func (h LsCxn) Help() string {
	return "Lists the connection addresses to the components of the system."
}

func (h LsCxn) Usage() string {
	return h.Help()
}

func (h LsCxn) Execute(shell *Shell, args []string) {
	dbcxn := config.GetDatabaseConnectionString()
	fmt.Printf("Database: %v\n", dbcxn)

	streamdb, _ := config.GetStreamdbDirectory()
	fmt.Printf("Streamdb: %v\n", streamdb)

	redis := config.GetConfiguration().GetRedisUri()
	fmt.Printf("Redis: %v\n", redis)

	gnatsd := config.GetConfiguration().GetGnatsdUri()
	fmt.Printf("Gnatsd: %v\n", gnatsd)
}

func (h LsCxn) Name() string {
	return "lscxn"
}
