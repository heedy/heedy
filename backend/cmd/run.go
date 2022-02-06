package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/server"
	"github.com/heedy/heedy/backend/updater"

	"github.com/spf13/cobra"
)

var createIfNotExists bool

var RunCmd = &cobra.Command{
	Use:   "run [location of database]",
	Short: "Runs existing heedy database",
	Long:  `Runs heedy using the passed database. If no folder is specifed, uses the default database location.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, err := GetDirectory(args)
		if err != nil {
			return err
		}
		c := assets.NewConfiguration()
		c.Verbose = verbose
		if loglevel != "" {
			c.LogLevel = &loglevel
		}
		if logdir != "" {
			c.LogDir = &logdir
		}

		if _, err := os.Stat(path.Join(directory, "heedy.conf")); os.IsNotExist(err) {
			// A heedy database does not exist in the config directory
			if !createIfNotExists {
				return fmt.Errorf("no database found at %s", directory)
			}

			return server.Setup(server.SetupContext{
				Config:    c,
				Directory: directory,
			}, ":1324")
		}

		if err = writepid(directory); err != nil {
			return err
		}
		defer delpid(directory)

		return updater.Run(updater.Options{
			ConfigDir:   directory,
			AddonConfig: c,
			Revert:      revert,
			Update:      applyUpdates,
			Runner: func(a *assets.Assets) error {
				return server.Run(a, nil)
			},
		})
	},
}

func init() {
	RunCmd.PersistentFlags().BoolVar(&createIfNotExists, "create", false, "Create the database if it doesn't exist")
	RootCmd.AddCommand(RunCmd)
}
