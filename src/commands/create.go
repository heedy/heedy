package commands

import (
	"config"
	"dbsetup"

	"errors"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Whether the new database should use testing configuration
	testConfiguration bool
	user              string
	email             string
)

// CreateCmd creates a new database
var CreateCmd = &cobra.Command{
	Use:   "create [location for database]",
	Short: "Create a new ConnectorDB database",
	Long: `Sets up the given directory with a new ConnectorDB database.
The directory must not exist. Make sure that no ConnectorDB instance is
running on default ports during setup, since server setup runs on default
configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Must specify directory in which to create database")
		}
		if len(args) > 1 {
			return ErrTooManyArgs
		}

		cfg := config.NewConfiguration()

		cfg.DatabaseDirectory = args[0]

		if testConfiguration {
			log.Warn("Test flag is set: Using testing configuration!")
			cfg = &config.TestConfiguration
		}

		if user != "" {
			usrpass := strings.Split(user, ":")
			if len(usrpass) != 2 {
				return errors.New("User must be in username:password format")
			}
			cfg.InitialUser = &config.UserMaker{
				Name:     usrpass[0],
				Password: usrpass[1],
				Email:    email,
				Role:     "admin",
			}
		}

		return dbsetup.Create(cfg)

	},
}

func init() {
	CreateCmd.Flags().BoolVar(&testConfiguration, "test", false, "Use testing configuration")
	CreateCmd.Flags().StringVar(&user, "user", "", "Admin user to create by default in username:password format")
	CreateCmd.Flags().StringVar(&email, "email", "root@localhost", "Email to use for the created admin user")

	RootCmd.AddCommand(CreateCmd)
}
