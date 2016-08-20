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

	mkredis    bool
	mkpostgres bool
	mkgnatsd   bool

	sqltype string
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

		dboptions := &dbsetup.Options{
			DatabaseDirectory: args[0],
			Config:            config.NewConfiguration(),

			RedisEnabled:  true,
			GnatsdEnabled: true,
			SQLEnabled:    true,
		}

		if testConfiguration {
			log.Warn("Test flag is set: Using testing configuration!")
			dboptions.Config = &config.TestConfiguration
			dboptions.InitialUser = config.TestUser
		}

		// Set up the database type
		if sqltype != "" {
			dboptions.Config.Sql.Type = sqltype
		}

		// If one of the create flags is given, we set enabled to the given values
		if mkredis || mkgnatsd || mkpostgres {
			log.Infof("Setting up: redis=%t nats=%t sql=%t (frontend is disabled)", mkredis, mkgnatsd, mkpostgres)
			dboptions.RedisEnabled = mkredis
			dboptions.GnatsdEnabled = mkgnatsd
			dboptions.SQLEnabled = mkpostgres

			dboptions.Config.Redis.Enabled = mkredis
			dboptions.Config.Sql.Enabled = mkpostgres
			dboptions.Config.Nats.Enabled = mkgnatsd

			// If the custom flags were given, we disable frontend, since it is assumed that
			// this is a power user who wants to run the backend over a cluster.
			dboptions.Config.Frontend.Enabled = false
		}

		if user != "" {
			usrpass := strings.Split(user, ":")
			if len(usrpass) != 2 {
				return errors.New("User must be in username:password format")
			}
			dboptions.InitialUser = &config.UserMaker{
				Name:     usrpass[0],
				Password: usrpass[1],
				Email:    email,
				Role:     "admin",
			}
		}

		// set up logging based on create config (this allows debug msgs in testing config)
		setLogging(dboptions.Config)

		return dbsetup.Create(dboptions)

	},
}

func init() {
	CreateCmd.Flags().BoolVar(&testConfiguration, "test", false, "Use testing configuration")
	CreateCmd.Flags().StringVar(&user, "user", "", "Admin user to create by default in username:password format")
	CreateCmd.Flags().StringVar(&email, "email", "root@localhost", "Email to use for the created admin user")

	CreateCmd.Flags().BoolVar(&mkredis, "redis", false, "set up the backend redis server (if no flags given, all created)")
	CreateCmd.Flags().BoolVar(&mkgnatsd, "nats", false, "set up the backend nats server (if no flags given, all created)")
	CreateCmd.Flags().BoolVar(&mkpostgres, "sql", false, "set up the backend sql server (if no flags given, all created)")

	CreateCmd.Flags().StringVar(&sqltype, "sqlbackend", "", "choose the backing server (postgres or sqlite3)")

	RootCmd.AddCommand(CreateCmd)
}
