package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop [location of database]",
	Short: "Stops heedy",
	Long:  `Shuts down heedy running in the background`,
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, err := GetDirectory(args)
		if err != nil {
			return err
		}

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
