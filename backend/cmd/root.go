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
	Use:   "heedy",
	Short: "Heedy is an IoT and QS repository and analysis system",
	Long:  `Heedy is a database built for interacting with your IoT devices and for storing your quantified-self data. It is buit to be extensible and self-contained, with a powerful plugin system allowing for in-depth analysis and action.`,
}

// Execute runs the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
