package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop [location of database]",
	Short: "Stops heedy",
	Long:  `Shuts down heedy running in the background`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		b, err := ioutil.ReadFile(path.Join(directory, "heedy.pid"))
		if err != nil {
			return err
		}
		pid, err := strconv.Atoi(string(b))
		if err != nil {
			return err
		}
		p, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		logrus.Infof("Sending SIGINT to %d", pid)
		return p.Signal(os.Interrupt)
	},
}

func init() {
	RootCmd.AddCommand(StopCmd)
}
