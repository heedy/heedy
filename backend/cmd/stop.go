package cmd

import (
	"os"

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

		p, err := getpid(directory)
		if err != nil {
			return err
		}
		logrus.Infof("Sending SIGINT to %d", p.Pid)
		return p.Signal(os.Interrupt)
	},
}

func init() {
	RootCmd.AddCommand(StopCmd)
}
