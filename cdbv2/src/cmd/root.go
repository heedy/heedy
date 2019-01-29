package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "connectordb",
	Short: "ConnectorDB is a repository for your quantified-self and IoT data",
	Long:  `ConnectorDB is a database built for interacting with your IoT devices and for storing your quantified-self data.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
