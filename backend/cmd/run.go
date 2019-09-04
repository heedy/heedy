package cmd

import (
	"encoding/json"
	"path/filepath"
	"path"
	"os"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/server"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var verbose bool

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
			if err!=nil {
				return err
			}
			directory = path.Join(f,"heedy")
		}
		var err error
		directory, err = filepath.Abs(directory)
		if err!=nil {
			return err
		}
		logrus.Infof("Using database at %s",directory)
		
		a, err := assets.Open(directory, nil)
		if err != nil {
			return err
		}
		assets.SetGlobal(a)

		b, err := json.MarshalIndent(a.Config, "", " ")
		if err != nil {
			return err
		}
		logrus.Debug(string(b))

		return server.Run(&server.RunOptions{
			Verbose: verbose,
		})
	},
}

func init() {
	RunCmd.Flags().BoolVar(&verbose, "verbose", false, "Extremely verbose logging of server requests and responses. Only works in DEBUG log level.")
	RootCmd.AddCommand(RunCmd)

}
