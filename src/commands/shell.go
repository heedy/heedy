package commands

import (
	"config"
	"connectordb"
	"shell"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ShellCmd runs the ConnectorDB shell
var ShellCmd = &cobra.Command{
	Use:   "shell [config file path or database directory] [optional: shell command]",
	Short: "Run commands on the ConnectorDB database",
	Long: `Runs the ConnectorDB shell, allowing you to interact with your
ConnectorDB database from the command line.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrConfig
		}
		cfg, err := config.LoadConfig(args[0])
		if err != nil {
			return err
		}

		setLogging(cfg)

		// Open the ConnectorDB database
		db, err := connectordb.Open(cfg.Options())
		if err != nil {
			return err
		}
		defer db.Close()

		if len(args) == 1 {
			shell.StartShell(db)
		} else {
			//We append all of the arguments, so that the shell gets the full thing
			scmd := args[1]
			for i := 2; i < len(args); i++ {
				scmd = scmd + " " + args[i]
			}
			log.Infof("Running '%s' in connectordb shell", scmd)
			s := shell.CreateShell(db)
			s.RunCommand(scmd)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(ShellCmd)
}
