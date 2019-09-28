package cmd

import (
	"os"
	"path"
	"path/filepath"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/server"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [location of database]",
	Short: "Runs heedy",
	Long:  `Runs heedy using the passed database. If no folder is specifed, uses the default database location.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return ErrTooManyArgs
		}
		var directory string
		if len(args) == 1 {
			directory = args[0]
		} else {
			f, err := os.UserConfigDir()
			if err != nil {
				return err
			}
			directory = path.Join(f, "heedy")
		}
		var err error
		directory, err = filepath.Abs(directory)
		if err != nil {
			return err
		}
		logrus.Infof("Using database at %s", directory)
		c := assets.NewConfiguration()
		c.Verbose = verbose
		a, err := assets.Open(directory, c)
		if err != nil {
			return err
		}
		assets.SetGlobal(a)

		return server.Run(a, nil)
	},
}

func init() {
	RootCmd.AddCommand(RunCmd)
}
