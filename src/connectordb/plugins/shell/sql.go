package shell

/**

Provides the ability to open the sql database locally.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"connectordb/config"
	"os"
	"os/exec"
)

// The clear command
type Sql struct {
}

func (h Sql) Help() string {
	return "Runs an interactive database shell"
}

func (h Sql) Usage() string {
	return "sql"
}

func (h Sql) Execute(shell *Shell, args []string) {

	cmd := exec.Command("psql", config.GetDatabaseConnectionString())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func (h Sql) Name() string {
	return "sql"
}
