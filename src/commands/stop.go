package commands

import (
	"dbsetup"
	"errors"
	"util"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// StopCmd stops the background servers
var StopCmd = &cobra.Command{
	Use:   "stop [config file path or database directory]",
	Short: "Stops ConnectorDB daemons",
	Long:  `Stops the servers started with the start command`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrConfig
		}
		if len(args) > 1 {
			return ErrTooManyArgs
		}

		if !util.IsDirectory(args[0]) {
			return errors.New("Could not find the given directory")
		}

		log.Info("Stopping ConnectorDB")

		return dbsetup.Stop(&dbsetup.Options{
			DatabaseDirectory: args[0],
		})

	},
}

func init() {
	RootCmd.AddCommand(StopCmd)
}
