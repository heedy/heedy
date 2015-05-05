package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"streamdb"
	"streamdb/dbmaker"
	"strings"
	"log"
	"streamdb/config"
	"streamdb/util"
	_ "plugins/shell"
	_ "plugins/webclient"
	"plugins"
)

var (
	createFlags            = flag.NewFlagSet("create", flag.ExitOnError)
	createUsernamePassword = createFlags.String("user", "admin:admin", "The initial user in username:password format")
	createEmail            = createFlags.String("email", "root@localhost", "The email address for the root user")
	createDbType           = createFlags.String("dbtype", "postgres", "The type of database to create.")

	startFlags  = flag.NewFlagSet("create", flag.ExitOnError)
	startBasic  = startFlags.Bool("basic", false, "Start only background servers")
	startRest   = startFlags.Bool("rest", true, "Start the REST API")
	startWriter = startFlags.Bool("dbwriter", true, "Start the databaseWriter")

	stopFlags = flag.NewFlagSet("stop", flag.ExitOnError)

	upgradeFlags = flag.NewFlagSet("upgrade", flag.ExitOnError)

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
)

//PrintUsage gives a nice message of the functionality available from the executable
func PrintUsage() {
	fmt.Printf("ConnectorDB Version %v\nCompiled for %v using %v\n\n", streamdb.Version, runtime.GOARCH, runtime.Version())
	fmt.Printf("Usage:\nconnectordb [command] [path to database folder] [--flags] \n")

	fmt.Printf("\ncreate: Initialize a new database at the given folder\n")
	createFlags.PrintDefaults()
	fmt.Printf("\nstart: Starts the given database\n")
	startFlags.PrintDefaults()
	fmt.Printf("\nstop: Shuts down all processes associated with the given database.\n")
	stopFlags.PrintDefaults()
	fmt.Printf("\nupgrade: Upgrades an existing database to a newer version.\n")
	upgradeFlags.PrintDefaults()
	fmt.Printf("\n")

	// Print all usages of the plugins
	plugins.Usage()

	fmt.Printf("\n")

}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// The main entrypoint into connectordb
func main() {

	// global system stuff
	flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }

        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

	// Make sure we don't go OOB
	if len(flag.Args()) < 2 {
		PrintUsage()
		return
	}

	// Choose our command
	var err error
	commandName := flag.Args()[0]
	dbPath      := flag.Args()[1]

	// Make sure this is abs.
	dbPath, _    = filepath.Abs(dbPath)
	log.Println(dbPath)
	// init and save later
	config.InitConfiguration(dbPath)
	defer config.SaveConfiguration()

	switch commandName {
		case "create":
			err = createDatabase()

		case "start":
			err = startDatabase(dbPath)

		case "stop":
			err = stopDatabase(dbPath)

		case "upgrade":
			err = upgradeDatabase(dbPath)

		default:
			err = runPlugin(commandName, dbPath)
			if err == plugins.ErrNoPlugin {
				PrintUsage()
				return
			}
	}

	if err != nil {
		fmt.Printf("Error: A problem occured during %v:\n\n%v\n", commandName, err)
	}
}


// processes the flags and makes sure they're valid, exiting if needed.
func processFlags(fs *flag.FlagSet) {
	fs.Parse(flag.Args()[2:])
}

// Does the creations step
func createDatabase() error {
	processFlags(createFlags)

	//extract the username and password from the formatted string
	usernamePasswordArray := strings.Split(*createUsernamePassword, ":")
	if len(usernamePasswordArray) != 2 {
		fmt.Println("--user: Username and password not given in format <username>:<password>")
		createFlags.PrintDefaults()
		return nil
	}
	username := usernamePasswordArray[0]
	password := usernamePasswordArray[1]

	config.GetConfiguration().DatabaseType = *createDbType
	log.Println(config.GetConfiguration())

	log.Println("CONNECTORDB: Doing Init")
	if err := dbmaker.Init(config.GetConfiguration()); err != nil {
		return err
	}

	log.Println("CONNECTORDB: Creating Files")
	if err := dbmaker.Create(config.GetConfiguration(), username, password, *createEmail); err != nil {
		return err
	}

	log.Println("CONNECTORDB: Stopping any subsystems")

	dbmaker.Stop(config.GetConfiguration())
	//dbmaker.Kill(config.GetConfiguration())

	fmt.Printf("\nDatabase created successfully.\n")
	return nil
}

func startDatabase(dbPath string) error {
	processFlags(startFlags)

	dbPath, err := util.ProcessConnectordbDirectory(dbPath)
	if err != nil {
		return err
	}

	if err := dbmaker.Init(config.GetConfiguration()); err != nil {
		return err
	}


	return dbmaker.Start(config.GetConfiguration())
}

func stopDatabase(dbPath string) error {
	processFlags(stopFlags)

	dbPath, err := util.ProcessConnectordbDirectory(dbPath)
	if err == nil {
		log.Printf("Connectordb looks like it isn't already running, but we'll try anyway.")
		return err
	}

	if err := dbmaker.Init(config.GetConfiguration()); err != nil {
		return err
	}

	if err := dbmaker.Stop(config.GetConfiguration()); err != nil {
		log.Printf("%v\n", err.Error())
	}

	return nil
}

func upgradeDatabase(dbPath string) error {
	processFlags(upgradeFlags)

	// get cannonicalized path and make sure we're not already running
	dbPath, err := util.ProcessConnectordbDirectory(dbPath)
	if err != nil {
		return err
	}

	// Start the server

 	return dbmaker.Upgrade()
}

func runPlugin(cmd, dbPath string) error {
	db, err := streamdb.OpenFromConfig(config.GetConfiguration())
	if err != nil {
		return err
	}
	defer db.Close()

	return plugins.Run(cmd, db, flag.Args()[2:])
}
