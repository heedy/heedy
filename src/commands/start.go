package commands

import (
	"dbsetup"
	"errors"
	"os"
	"path/filepath"
	"util"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// whether to force a start even if pid files exist
	force bool
)

// StartCmd starts the background servers
var StartCmd = &cobra.Command{
	Use:   "start [config file path or database directory]",
	Short: "Starts ConnectorDB's backend databases",
	Long: `ConnectorDB uses postgres, redis, and gnatsd in the background.
This command starts these services, so that they run in the background.`,
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

		log.Info("Starting Database")

		//force removes the pid file
		if force {
			log.Warn("Force Flag: removing connectordb.pid")
			os.Remove(filepath.Join(args[0], "connectordb.pid"))
		}

		err := dbsetup.Start(args[0])

		if err == dbsetup.ErrAlreadyRunning {
			log.Error(err.Error())
			return errors.New("Use the --force flag if you know that it is not.")
		}
		return err
	},
}

func init() {
	StartCmd.Flags().BoolVar(&force, "force", false, "forces start of the database even if connectordb.pid exists")
	RootCmd.AddCommand(StartCmd)
}
