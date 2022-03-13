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
	configFile string
	addr       string
	username   string
	password   string
	plugins    []string
)

// CreateCmd creates a new database
var CreateCmd = &cobra.Command{
	Use:   "create [location to put database]",
	Short: "Create a new database at the specified location",
	Long: `Sets up the given directory with a new heedy database.
Creates the folder if it doesn't exist, but fails if the folder is not empty. If no folder is specified, uses the default database location.

If you want to set it up from command line, without the setup server, you can specify a configuration file and the admin user:
   
  heedy create ./myfolder --noserver -c ./heedy.conf --username=myusername --password=mypassword

It is recommended that new users use the web setup, which will guide you in preparing the database for use:

  heedy create ./myfolder

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, err := GetDirectory(args)
		if err != nil {
			return err
		}

		c := assets.NewConfiguration()
		c.Addr = &addr
		c.Verbose = verbose
		if loglevel != "" {
			c.LogLevel = &loglevel
		}
		if logdir != "" {
			c.LogDir = &logdir
		}
		if len(plugins) > 0 {
			plugin_names := make([]string, len(plugins))
			for i, p := range plugins {
				plugin_names[i] = path.Base(p)
			}
			c.ActivePlugins = &plugin_names
		}

		sc := server.SetupContext{
			CreateOptions: assets.CreateOptions{
				Config:     c,
				Directory:  directory,
				ConfigFile: configFile,
				Plugins:    plugins,
			},
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
		return server.Setup(sc)

	},
}

func init() {
	CreateCmd.Flags().StringVar(&addr, "addr", ":1324", "The address at which to run heedy")
	CreateCmd.Flags().BoolVar(&noserver, "noserver", false, "Don't start the setup server - directly create the database.")
	CreateCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to an existing configuration file to use for the database.")
	CreateCmd.Flags().StringVar(&username, "username", "", "Default user's username")
	CreateCmd.Flags().StringVar(&password, "password", "", "Default user's password")
	CreateCmd.Flags().StringVar(&testapp, "testapp", "", "Whether to create a test app with the given access token. Only works in noserver mode")
	CreateCmd.Flags().StringSliceVarP(&plugins, "plugin", "p", []string{}, "A plugin folder to auto-enable")

	RootCmd.AddCommand(CreateCmd)
}
