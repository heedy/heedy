package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (

	// ErrTooManyArgs is called when given too many args
	ErrTooManyArgs = errors.New("Too many arguments were specified.")
)

// RootCmd is the root command under which all other commands are placed.
// It is used to initialize all variables that are global for the whole app
var RootCmd = &cobra.Command{
	Use:   "connectordb",
	Short: "ConnectorDB is a repository for your quantified-self and IoT data",
	Long:  `ConnectorDB is a database built for interacting with your IoT devices and for storing your quantified-self data.`,
	Run: func(cmd *cobra.Command, args []string) {
		//server.RunServer()
		//cmd.HelpFunc()(cmd, args)
	},
}

// Execute runs the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
