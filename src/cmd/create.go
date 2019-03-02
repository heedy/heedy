package cmd

import (
	"os"

	"github.com/connectordb/connectordb/src/assets"

	"github.com/connectordb/connectordb/src/database"
	"github.com/connectordb/connectordb/src/server"

	"github.com/spf13/cobra"
)

var (

	// Should we run the setup UI?
	nosetup    bool
	setupHost  string
	configFile string
	port       uint16
	host       string
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
		if len(args) > 1 {
			return ErrTooManyArgs
		}

		directory := ""
		if len(args) == 1 {
			directory = args[0]
		}
		c := assets.NewConfiguration()
		if port != 0 {
			c.Port = &port
		}
		if host != "_" {
			c.Host = &host
		}

		if nosetup {
			a, err := assets.Create(directory, c, configFile)
			if err != nil {
				return err
			}
			if err = database.Create(a); err != nil {
				os.RemoveAll(directory)
				return err
			}
			return nil
		}
		return server.Setup(directory, c, configFile, setupHost)

	},
}

func init() {
	CreateCmd.Flags().Uint16VarP(&port, "port", "p", 0, "The port on which to run ConnectorDB")
	CreateCmd.Flags().StringVar(&host, "host", "_", "The host on which to run ConnectorDB")
	CreateCmd.Flags().BoolVar(&nosetup, "nosetup", false, "Don't start the setup server - directly create the database.")
	CreateCmd.Flags().StringVar(&setupHost, "setup", ":8000", "Run a setup server on the given host:port")
	CreateCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to an existing configuration file to use for the database.")

	RootCmd.AddCommand(CreateCmd)
}
