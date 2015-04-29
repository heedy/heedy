package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"streamdb"
	"streamdb/dbmaker"
	"strings"
	"log"
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
	fmt.Printf("Usage:\nconnectordb [command] [--flags] [path to database folder]\n")

	fmt.Printf("\ncreate: Initialize a new database at the given folder\n")
	createFlags.PrintDefaults()
	fmt.Printf("\nstart: Starts the given database\n")
	startFlags.PrintDefaults()
	fmt.Printf("\nstop: Shuts down all processes associated with the given database.\n")
	stopFlags.PrintDefaults()
	fmt.Printf("\nupgrade: Upgrades an existing database to a newer version.\n")
	upgradeFlags.PrintDefaults()

	fmt.Printf("\n")

}

// The main entrypoint into connectordb
func main() {
	if len(os.Args) < 3 {
		PrintUsage()
		return
	}

	flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

	log.Print(flag.Args())


	switch flag.Args()[0] {
	case "create":
		assertFlagsGood(createFlags)
		
		//extract the username and password from the formatted string
		usernamePasswordArray := strings.Split(*createUsernamePassword, ":")
		if len(usernamePasswordArray) != 2 {
			fmt.Println("--user: Username and password not given in format <username>:<password>")
			createFlags.PrintDefaults()
			return
		}
		username := usernamePasswordArray[0]
		password := usernamePasswordArray[1]

		err := dbmaker.Create(createFlags.Arg(0), *createDbType, nil)
		err = dbmaker.MakeUser(createFlags.Arg(0), username, password, *createEmail, err)

		if err != nil {
			fmt.Printf("Database creation FAILED with the following error:\n\n%v\n", err)
		} else {
			fmt.Printf("\nDatabase created successfully.\n")
		}

	case "start":
		assertFlagsGood(startFlags)

		//TODO: Load ports and interface from a config file
		err := dbmaker.Start(startFlags.Arg(0), "127.0.0.1", 6379, 4222, 52592, nil)
		if err != nil {
			fmt.Printf("ConnectorDB crashed with the following error:\n\n%v\n", err)
		}

	case "stop":
		assertFlagsGood(stopFlags)

		err := dbmaker.Stop(stopFlags.Arg(0), nil)
		if err != nil {
			fmt.Printf("ConnectorDB stop failed with the following error:\n\n%v\n", err)
		}

	case "upgrade":
		assertFlagsGood(upgradeFlags)

		err := dbmaker.Upgrade(upgradeFlags.Arg(0), nil)
		if err != nil {
			fmt.Printf("ConnectorDB upgrade failed with the following error:\n\n%v\n", err)
		}

	default:
		PrintUsage()
	}
}

func assertFlagsGood(fs *flag.FlagSet) {
	fs.Parse(flag.Args()[1:])
	if fs.NArg() != 1 {
		PrintUsage()
		os.Exit(1)
	}
}
