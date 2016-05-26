/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package main

import (
	"config"
	"connectordb"
	"dbsetup"
	"os"
	"path/filepath"
	"runtime/pprof"
	"server"
	"shell"
	"strings"
	"util"

	"github.com/codegangsta/cli"

	log "github.com/Sirupsen/logrus"

	pconfig "config/permissions"
)

func getDatabase(c *cli.Context) string {
	n := c.Args().First()
	if n == "" {
		log.Fatal("You must specify a directory or configuration file to use")
	}
	return n
}

func getConfiguration(c *cli.Context) *config.Configuration {
	//There are a few different situations that we handle here:
	//1) A database folder is given
	//		In this case we read the internal connectordb.conf file to get the config
	//2) A config file is given
	//		We read the file
	var err error
	arg := getDatabase(c)

	if util.IsDirectory(arg) {
		arg = filepath.Join(arg, "connectordb.conf")
	}
	err = config.SetPath(arg)
	if err != nil {
		log.Fatal(err.Error())
	}

	//Print out the configuration as we understand it
	log.Debug(config.Get().String())

	return config.Get()
}

func runconfigCallback(c *cli.Context) error {
	n := c.Args().First()
	if n == "" {
		log.Fatal("You must specify the file to write config to")
	}

	cfg := config.NewConfiguration()
	err := cfg.Save(n)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

func runpermissionsCallback(c *cli.Context) error {
	n := c.Args().First()
	if n == "" {
		log.Fatal("You must specify the file to write permissions to")
	}

	err := pconfig.Default.Save(n)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

func runConnectorDBCallback(c *cli.Context) error {
	cfg := getConfiguration(c)

	//The run command allows to set the host and port to run server on
	cfg.Hostname = c.String("host")
	cfg.Port = uint16(c.Int("port"))
	// Enable random people to join
	if c.Bool("join") {
		pconfig.Get().UserRoles["nobody"].Join = true
	}
	if c.Bool("http") {
		log.Info("Running in http-only mode")
		cfg.TLS.Key = ""
		cfg.TLS.Cert = ""
	}

	err := server.RunServer()
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

func runShellCallback(c *cli.Context) error {
	cfg := getConfiguration(c)
	db, err := connectordb.Open(cfg.Options())
	defer db.Close()
	if err != nil {
		log.Fatal(err.Error())
	}

	scmd := c.String("exec")
	if scmd == "" {
		shell.StartShell(db)
	} else {
		s := shell.CreateShell(db)
		s.RunCommand(scmd)
	}

	return nil
}

//This is called when the user runs "connectordb create"
func createDatabaseCallback(c *cli.Context) error {
	cfg := config.NewConfiguration()
	if c.Bool("test") {
		log.Warn("test flag: Using testing configuration!")
		cfg = &config.TestConfiguration
	}

	cfg.DatabaseDirectory = getDatabase(c)

	//Next we parse the user flags
	uname := c.String("user")
	if uname != "" {
		usrpass := strings.Split(uname, ":")
		if len(usrpass) != 2 {
			log.Fatal("The username flag must be in username:password format")
		}
		cfg.InitialUser = &config.UserMaker{
			Name:     usrpass[0],
			Password: usrpass[1],
			Email:    c.String("email"),
			Role:     "admin",
		}
	}

	err := dbsetup.Create(cfg)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

//This is called when the user runs "connectordb start"
func startDatabaseCallback(c *cli.Context) error {
	log.Info("Starting Database")
	//force removes the pid file
	if c.Bool("force") {
		os.Remove(filepath.Join(getDatabase(c), "connectordb.pid"))
	}

	err := dbsetup.Start(getDatabase(c))
	if err != nil {
		log.Error(err.Error())
		if err == dbsetup.ErrAlreadyRunning {
			log.Error("Use the --force flag if you know that it is not.")
		}
	}
	return err
}

//This is called when the user runs "connectordb stop"
func stopDatabaseCallback(c *cli.Context) error {
	log.Info("Stopping Database")

	err := dbsetup.Stop(getDatabase(c))
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

func main() {

	app := cli.NewApp()
	app.Name = "ConnectorDB"
	app.Usage = "Run and administer a ConnectorDB database."
	app.Version = connectordb.Version

	app.Copyright = "This software is available under the MIT license."
	app.Authors = []cli.Author{{Name: "ConnectorDB team", Email: "support@connectordb.com"}}

	app.Commands = []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create a new ConnectorDB database",
			Action:  createDatabaseCallback,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Value: "",
					Usage: "The admin user to create by default in username:password format",
				},
				cli.StringFlag{
					Name:  "email",
					Value: "root@localhost",
					Usage: "The email to use for the created admin user",
				},
				cli.BoolFlag{
					Name:  "test",
					Usage: "Use the special test configuration for the database",
				},
			},
		},
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "Start ConnectorDB's backend databases",
			Action:  startDatabaseCallback,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force",
					Usage: "forces start of the database even if connectordb.pid exists",
				},
			},
		},
		{
			Name:    "stop",
			Aliases: []string{"q"},
			Usage:   "Stop ConnectorDB's backend databases",
			Action:  stopDatabaseCallback,
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Run the ConnectorDB frontend server",
			Action:  runConnectorDBCallback,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "host",
					Value: "",
					Usage: "The interface to which the ConnectorDB server should bind",
				},
				cli.IntFlag{
					Name:  "port, p",
					Value: 8000,
					Usage: "The port on which to run the ConnectorDB server",
				},
				cli.BoolFlag{
					Name:  "http",
					Usage: "forces server to run in http mode even when TLS cert/key are in conf",
				},
				cli.BoolFlag{
					Name:  "join",
					Usage: "Enables free join on the server (anyone can join)",
				},
			},
		},
		{
			Name:    "shell",
			Aliases: []string{},
			Usage:   "Runs an administrative shell on the database",
			Action:  runShellCallback,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exec, e",
					Value: "",
					Usage: "Instead of running connectordb shell in interactive mode, execute the given commands",
				},
			},
		},
		{
			Name:   "config",
			Usage:  "Creates a new configuration file with defaults at the given path.",
			Action: runconfigCallback,
		},
		{
			Name:   "permissions",
			Usage:  "Creates a new configuration file with default permissions at the given path.",
			Action: runpermissionsCallback,
		},
	}

	//Set up the global flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lvl, l",
			Value: "INFO",
			Usage: "The log level to use when logging (DEBUG,INFO,WARN,ERROR)",
		},
		cli.StringFlag{
			Name:  "log",
			Value: "",
			Usage: "Log file to write",
		},
		cli.StringFlag{
			Name:  "cpuprof",
			Value: "",
			Usage: "Write a CPU profile of the application",
		},
	}

	var cpufile *os.File
	var logfile *os.File
	var err error

	//Initialize global environment before running anything
	app.Before = func(c *cli.Context) error {

		//Set up the log level
		switch c.GlobalString("lvl") {
		default:
			log.Fatalln("Unrecognized log level ", c.GlobalString("lvl"), " must be one of DEBUG,INFO,WARN,ERROR.")
		case "INFO":
			log.SetLevel(log.InfoLevel)
		case "WARN":
			log.SetLevel(log.WarnLevel)
		case "DEBUG":
			log.SetLevel(log.DebugLevel)
			log.Debug("Setting DEBUG log level")
		case "ERROR":
			log.SetLevel(log.ErrorLevel)
		}

		//Set up the logfile
		logf := c.GlobalString("log")
		if logf != "" {
			log.Infof("Writing logs to %s", logf)
			logfile, err = os.OpenFile(logf, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("Could not open file %s: %s", logf, err.Error())
			}
			log.SetFormatter(new(log.JSONFormatter))
			log.SetOutput(logfile)
		}

		//Set up CPU profiling if it is enabled
		cpuf := c.GlobalString("cpuprof")
		if cpuf != "" {
			log.Debug("Writing CPU profile to ", cpuf)

			cpufile, err = os.Create(cpuf)
			if err != nil {
				log.Fatal(err)
			}

			pprof.StartCPUProfile(cpufile)

			//Now, the program will be closed in some way, which might be hard to catch. We use
			//close on exit to shut down the CPU profile
			util.CloseOnExit(util.CloseCall{func() {
				log.Debug("Writing CPU profile...")
				pprof.StopCPUProfile()
				cpufile.Close()
			}})

		}

		return nil
	}

	app.Run(os.Args)
}
