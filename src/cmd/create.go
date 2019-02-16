package cmd

import (
	"errors"

	"github.com/connectordb/connectordb/assets"
	"github.com/spf13/afero"

	"github.com/spf13/cobra"
)

var (

	// Should we run the setup UI?
	nosetup bool
	setup   string
)

// CreateCmd creates a new database
var CreateCmd = &cobra.Command{
	Use:   "create [location to put database]",
	Short: "Create a new database",
	Long: `Sets up the given directory with a new ConnectorDB database.
Creates the folder if it doesn't exist, but fails if the folder is not empty.

If you want to set it up from command line, without the setup server, you can specify a configuration file and the admin user:
   
  connectordb create ./myfolder -c ./connectordb.conf --user=myusername --password=mypassword

It is recommended that new users use the web setup, which will guide you in preparing the database for use:

  connectordb create ./myfolder

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Must specify directory in which to create database")
		}
		if len(args) > 1 {
			return ErrTooManyArgs
		}

		// Setting up the database: first we load the assets
		assetFs := assets.BuiltinAssets()
		osFs := afero.NewOsFs()
		err := assets.CopyDir(assetFs, "/newdb", osFs, args[0])
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	CreateCmd.Flags().StringVar(&setup, "setup", ":8000", "Start a setup server on the given host:port")

	RootCmd.AddCommand(CreateCmd)
}
