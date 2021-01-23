package cmd

import (
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/server"
	"github.com/heedy/heedy/backend/updater"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [location of database]",
	Short: "Runs heedy",
	Long:  `Runs heedy using the passed database. If no folder is specifed, uses the default database location.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, err := GetDirectory(args)
		if err != nil {
			return err
		}
		c := assets.NewConfiguration()
		c.Verbose = verbose

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
	RootCmd.AddCommand(RunCmd)
}
