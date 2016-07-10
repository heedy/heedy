package commands

import (
	"connectordb"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"util"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// The git hash/buildstamp is inserted into the build during link time
	// http://www.atatus.com/blog/golang-auto-build-versioning/
	GitHash    = ""
	BuildStamp = ""

	// ErrConfig is shown when a configuration is expected for ConnectorDB
	ErrConfig      = errors.New("You must specify a database directory or config file")
	ErrTooManyArgs = errors.New("Too many arguments were specified.")

	loglevel   string
	logfile    string
	cpuprofile string
	version    bool
)

// RootCmd is the root command under which all other commands are placed.
// It is used to initialize all ariables that are global for the whole app
var RootCmd = &cobra.Command{
	Use:   "connectordb",
	Short: "ConnectorDB is a repository for your quantified-self and IoT data",
	Long: `ConnectorDB is a quick and powerful database
built for interacting with your IoT devices and for
storing your quantified-self data.`,

	Run: func(cmd *cobra.Command, args []string) {
		if version {
			fmt.Printf("ConnectorDB %s\n\narch: %s/%s\ngo: %s\ngit: %s\nbuild: %s\n", connectordb.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), GitHash, BuildStamp)
		} else {
			cmd.Help()
		}
	},

	// Set up logging and profiling - everything that is needed for all runs of ConnectorDB
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		// First, set up the log level
		switch loglevel {
		default:
			return fmt.Errorf("Unrecognized log level %s. Must be one of debug,info,warn,error", loglevel)
		case "INFO", "info":
			log.SetLevel(log.InfoLevel)
		case "WARN", "warn":
			log.SetLevel(log.WarnLevel)
		case "DEBUG", "debug":
			log.SetLevel(log.DebugLevel)
			log.Debug("Setting DEBUG log level")
		case "ERROR", "error":
			log.SetLevel(log.ErrorLevel)
		}

		//  Next set up the log file
		if logfile != "" {
			log.Infof("Writing logs to %s", logfile)
			logf, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return fmt.Errorf("Could not open file %s: %s", logfile, err.Error())
			}
			log.SetFormatter(new(log.JSONFormatter))
			log.SetOutput(logf)
		}

		//Set up CPU profiling if it is enabled
		if cpuprofile != "" {
			log.Debug("Writing CPU profile to ", cpuprofile)

			cpufile, err := os.Create(cpuprofile)
			if err != nil {
				return err
			}

			pprof.StartCPUProfile(cpufile)

			//Now, the program will be closed in some way, which might be hard to catch. We use
			//close on exit to shut down the CPU profile.
			// TODO: CloseOnExit only is called on sigint/sigterm. It isn't called
			// on normal exit.
			util.CloseOnExit(util.CloseCall{Callme: func() {
				log.Debug("Writing CPU profile...")
				pprof.StopCPUProfile()
				cpufile.Close()
			}})

		}

		return nil
	},
}

func init() {

	RootCmd.PersistentFlags().StringVar(&logfile, "log", "", "The file to which log output is written")
	RootCmd.PersistentFlags().StringVarP(&loglevel, "loglevel", "l", "info", "The types of messages to show (debug,info,warn,error)")
	RootCmd.PersistentFlags().StringVar(&cpuprofile, "cpuprof", "", "File to which a cpu profile of ConnectorDB will be written")

	RootCmd.Flags().BoolVar(&version, "version", false, "Show ConnectorDB version and exit")
}
