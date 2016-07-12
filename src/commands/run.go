package commands

import (
	"config"
	"config/permissions"
	"path/filepath"
	"server"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Look at the flag declarations below for explanation of these variables
	host string
	port uint16
	http bool
	join bool
)

// RunCmd runs the ConnectorDB server.
var RunCmd = &cobra.Command{
	Use:   "run [config file path or database directory]",
	Short: "Runs the ConnectorDB frontend server",
	Long: `The ConnectorDB frontend server requires background servers which
are started with the start command. The frontend is
the main server that is exposed to the internet and runs the
ConnectorDB API and web app.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrConfig
		}
		if len(args) > 1 {
			return ErrTooManyArgs
		}

		cfg, err := config.LoadConfig(args[0])
		if err != nil {
			return err
		}

		setLogging(cfg)

		// Override the flag-based options if they are set
		// WARNING: These overrides will NOT hold if the config file is reloaded
		// during runtime.
		if host != "" {
			cfg.Hostname = host
		}
		if port != 0 {
			cfg.Port = port
		}
		if join {
			log.Warn("Enabling join for 'nobody' role. Anyone visiting server can create user.")
			// This overrides the nobody role to allow joining
			permissions.Get().UserRoles["nobody"].Join = true
		}
		if http {
			log.Info("Running in http mode")
			cfg.TLS.Key = ""
			cfg.TLS.Cert = ""
		}

		// Since the configuration changed, we need to revalidate it.
		// Note that we can't revalidate if file names have changed, since file names
		// are queried from the config file's directory, and we don't have exact location
		// at this point.
		// The reason this is OK is because we didn't change file locations, and Previous
		// validation already changed files to absolute paths.
		if err = cfg.Validate(); err != nil {
			return err
		}

		// Print out the configuration as we understand it
		log.Debug(cfg.String())

		// The configuration is loaded globally. We now start the server, which
		// will use the config for its options.
		return server.RunServer()

	},
}

// setRunFlags is used to set up the flags used in start - it basically regenerates the flags so that
// when start is called with these flags, the connectordb run command will have the flags passed directly
// to it.
// That is, if I run `connectordb start testdb --join`, when starting the frontend, start will run:
// connectordb run testdb --join
// as if the flag were passed directly
func setRunFlags() []string {
	ret := []string{}
	// If there was a log file set, we set that up here
	if loglevel != "" {
		ret = append(ret, "--loglevel", loglevel)
	}
	if logfile != "" {
		var err error
		logfile, err = filepath.Abs(logfile)
		if err != nil {
			panic(err)
		}
		ret = append(ret, "--logfile", logfile)
	}
	if host != "" {
		ret = append(ret, "--host", host)
	}
	if port != 0 {
		ret = append(ret, "--port", strconv.Itoa(int(port)))
	}
	if http {
		ret = append(ret, "--http")
	}
	if join {
		ret = append(ret, "--join")
	}
	return ret
}

func init() {
	RunCmd.Flags().StringVar(&host, "host", "", "Override the interface to which the ConnectorDB server should bind")
	RunCmd.Flags().Uint16VarP(&port, "port", "p", 0, "Override the port on which to run frontend")
	RunCmd.Flags().BoolVar(&http, "http", false, "forces server to run in http mode even when TLS cert/key are in conf")
	RunCmd.Flags().BoolVar(&join, "join", false, "Enables free join on the server (anyone can join)")

	// These commands are for mirroring join's flags in start. See setRunFlags
	StartCmd.Flags().StringVar(&host, "host", "", "Override the interface to which the ConnectorDB server should bind")
	StartCmd.Flags().Uint16VarP(&port, "port", "p", 0, "Override the port on which to run frontend")
	StartCmd.Flags().BoolVar(&http, "http", false, "forces server to run in http mode even when TLS cert/key are in conf")
	StartCmd.Flags().BoolVar(&join, "join", false, "Enables free join on the server (anyone can join)")

	RootCmd.AddCommand(RunCmd)
}
