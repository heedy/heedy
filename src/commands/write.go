package commands

import (
	"errors"

	"config"
	"config/permissions"

	"github.com/spf13/cobra"
)

// WriteCmd allows writing specific files from ConnectorDB config
var WriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Create specific configuration files",
	Long: `The write command is used to write default versions
of specific configuration files to a given location.
This can be used to set up special permissions or replace/upgrade
files within an existing database.`,
}

// WritePermissions writes the permissions file
var WritePermissions = &cobra.Command{
	Use:   "permissions [filename]",
	Short: "Create a ConnectorDB permissions file",
	Long: `The permissions file contains the full permissions matrix
used for specifying the exact abilities/permissions each role
has in the database. This command creates a permissions file
with the default permissions setup that you can modify to suit
your needs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You must specify a file to write permissions to")
		}
		if len(args) > 1 {
			return errors.New("Found extra parameter - expecting only a filename")
		}

		return permissions.Default.Save(args[0])
	},
}

// WriteConfig writes a configuration file
var WriteConfig = &cobra.Command{
	Use:   "config [filename]",
	Short: "Create a ConnectorDB configuration file",
	Long: `The configuration file is passed into ConnectorDB when running.
It contains all of the information necessary to run the
database, and all supported database parameters`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You must specify a file to write config to")
		}
		if len(args) > 1 {
			return errors.New("Found extra parameter - expecting only a filename")
		}
		return config.NewConfiguration().Save(args[0])
	},
}

func init() {
	WriteCmd.AddCommand(WritePermissions)
	WriteCmd.AddCommand(WriteConfig)

	RootCmd.AddCommand(WriteCmd)
}
