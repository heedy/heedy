package cmd

import (
	"errors"
	"os"
	"path"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/server"

	"github.com/spf13/cobra"
)

var (

	// Should we run the setup UI?
	noserver   bool
	testapp    string
	setupHost  string
	configFile string
	addr       string
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
			f, err := UserDataDir()
			if err != nil {
				return err
			}
			directory = path.Join(f, "heedy")
		}
		c := assets.NewConfiguration()
		if addr != "" {
			c.Addr = &addr
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
			err := server.SetupCreate(sc)
			if err == nil && testapp != "" {
				// If testapp is set, auto-creates an app with the given access token
				// This is specifically for creating a testing database that can directly
				// be accessed by API

				db, err := database.Open(assets.Get())
				if err != nil {
					os.RemoveAll(directory)
					return err
				}
				defer db.Close()
				appname := "Test App"
				appid, _, err := db.CreateApp(&database.App{
					Details: database.Details{
						Name: &appname,
					},
					Owner: &sc.User.UserName,
					Scope: &database.AppScopeArray{
						ScopeArray: database.ScopeArray{
							Scope: []string{"*"},
						},
					},
				})
				if err != nil {
					os.RemoveAll(directory)
					return err
				}
				// Manually set the access token to the value
				_, err = db.Exec("UPDATE apps SET access_token=? WHERE id=?;", testapp, appid)
				return err
			}
			return err
		} else if testapp != "" {
			return errors.New("testapp can only be set in noserver mode")
		}
		return server.Setup(sc, addr)

	},
}

func init() {
	CreateCmd.Flags().StringVar(&addr, "addr", ":1324", "The address at which to run heedy")
	CreateCmd.Flags().BoolVar(&noserver, "noserver", false, "Don't start the setup server - directly create the database.")
	CreateCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to an existing configuration file to use for the database.")
	CreateCmd.Flags().StringVar(&username, "username", "", "Default user's username")
	CreateCmd.Flags().StringVar(&password, "password", "", "Default user's password")
	CreateCmd.Flags().StringVar(&testapp, "testapp", "", "Whether to create a test app with the given access token. Only works in noserver mode")

	RootCmd.AddCommand(CreateCmd)
}
