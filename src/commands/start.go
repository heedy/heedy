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

	// The run matrix - which services to start
	runpostgres bool
	rungnatsd   bool
	runredis    bool
	runfrontend bool
	runbackend  bool
)

// StartCmd starts the background servers
var StartCmd = &cobra.Command{
	Use:   "start [config file path or database directory]",
	Short: "Starts ConnectorDB and its backend services as daemons",
	Long: `ConnectorDB uses postgres (or sqlite3), redis, and gnatsd in the background, and runs its own
frontend server using these services. The start command allows you to start
all of the services including the frontend as a daemon. You can also use the start command
with the --backend flag to start only the background servers, and can use connectordb run to
run the frontend server in the foreground. Don't forget to shut down the backend with
connectordb stop.`,
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

		log.Info("Starting ConnectorDB")

		//force removes the pid file
		if force {
			log.Warn("Force Flag: removing connectordb.pid")
			os.Remove(filepath.Join(args[0], "connectordb.pid"))
		}

		opt := &dbsetup.Options{
			DatabaseDirectory: args[0],
			RedisEnabled:      true,
			GnatsdEnabled:     true,
			SQLEnabled:        true,
			FrontendEnabled:   true,
			FrontendFlags:     setRunFlags(),
			FrontendPort:      port, // The port should be 0 by default. This is so waiting for port Open doesn't fail
		}

		// If any of the flags is given, we do manual start
		if runredis || rungnatsd || runpostgres || runfrontend || runbackend {
			if runbackend {
				runredis = true
				rungnatsd = true
				runpostgres = true
			}

			log.Infof("Starting: redis=%t sql=%t nats=%t frontend=%t", runredis, rungnatsd, runpostgres, runfrontend)

			opt.RedisEnabled = runredis
			opt.GnatsdEnabled = rungnatsd
			opt.SQLEnabled = runpostgres
			opt.FrontendEnabled = runfrontend
		}

		err := dbsetup.Start(opt)

		if err == dbsetup.ErrAlreadyRunning {
			log.Error(err.Error())
			return errors.New("Use the --force flag if you know that it is not.")
		}
		return err
	},
}

func init() {
	StartCmd.Flags().BoolVar(&force, "force", false, "forces start of the database even if connectordb.pid exists")

	StartCmd.Flags().BoolVar(&runredis, "redis", false, "start the backend redis server (if no flags given, all started)")
	StartCmd.Flags().BoolVar(&rungnatsd, "nats", false, "start the backend nats server (if no flags given, all started)")
	StartCmd.Flags().BoolVar(&runpostgres, "sql", false, "start the backend sql server (postgres) (if no flags given, all started)")
	StartCmd.Flags().BoolVar(&runfrontend, "frontend", false, "start the ConnectorDB server (same as connectordb run but as daemon) (if no flags given, all started)")
	StartCmd.Flags().BoolVar(&runbackend, "backend", false, "start only backend servers - same as --redis --nats --sql (if no flags given, all started)")

	RootCmd.AddCommand(StartCmd)
}
