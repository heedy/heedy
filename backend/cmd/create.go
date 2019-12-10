package cmd

import (
	"os"
	"path"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/server"

	"github.com/spf13/cobra"
)

var (

	// Should we run the setup UI?
	noserver   bool
	setupHost  string
	configFile string
	port       uint16
	host       string
	username   string
	password   string
)

// CreateCmd creates a new database
var CreateCmd = &cobra.Command{
	Use:   "create [location to put database]",
	Short: "Create a new database",
	Long: `Sets up the given directory with a new heedy database.
Creates the folder if it doesn't exist, but fails if the folder is not empty. If no folder is specified, uses the default database location.

If you want to set it up from command line, without the setup server, you can specify a configuration file and the admin user:
   
  heedy create ./myfolder --noserver -c ./heedy.conf --username=myusername --password=mypassword

It is recommended that new users use the web setup, which will guide you in preparing the database for use:

  heedy create ./myfolder

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return ErrTooManyArgs
		}

		var directory string
		if len(args) == 1 {
			directory = args[0]
		} else if len(args) == 0 {
			f, err := os.UserConfigDir()
			if err != nil {
				return err
			}
			directory = path.Join(f, "heedy")
		}
		c := assets.NewConfiguration()
		if port != 0 {
			c.Port = &port
		}
		if host != "_" {
			c.Host = &host
		}
		c.Verbose = verbose

		sc := server.SetupContext{
			Config:    c,
			Directory: directory,
			File:      configFile,
			User: server.SetupUser{
				UserName: username,
				Password: password,
			},
		}

		if noserver {
			return server.SetupCreate(sc)
		}
		return server.Setup(sc, setupHost)

	},
}

func init() {
	CreateCmd.Flags().Uint16VarP(&port, "port", "p", 0, "The port on which to run heedy")
	CreateCmd.Flags().StringVar(&host, "host", "_", "The host on which to run heedy")
	CreateCmd.Flags().BoolVar(&noserver, "noserver", false, "Don't start the setup server - directly create the database.")
	CreateCmd.Flags().StringVar(&setupHost, "setup", ":1324", "Run a setup server on the given host:port")
	CreateCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to an existing configuration file to use for the database.")
	CreateCmd.Flags().StringVar(&username, "username", "", "Default user's username")
	CreateCmd.Flags().StringVar(&password, "password", "", "Default user's password")

	RootCmd.AddCommand(CreateCmd)
}
