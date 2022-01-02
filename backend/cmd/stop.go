package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop [location of database]",
	Short: "Stops a heedy server running in the background",
	Long:  `Shuts down heedy running in the background, the main way to stop servers started using 'heedy start'.`,
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
