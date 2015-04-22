package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"plugins/rest"
	"streamdb"
	"streamdb/config"
	"streamdb/dbutil"
	"streamdb/users"
	"strings"
	"time"
	"syscall"
	"os/signal"

	"github.com/kardianos/osext"
	"github.com/vharitonsky/iniflags"
)

//ProgramUsage is a string describing how to use the program
const ProgramUsage = `

Usage: connectordb (create|start|stop|shell) directory

directory - the directory holding the configuration and the data for connectordb

create:
    Sets up a new connectordb instance at the given directory.

start:
    Starts connectordb and needed processes running from the given connectordb
    instance directory.

    Subcommands include:
        [blank] - default; start everything
        servers - redis, gnatsd, and postgres if required
        redis   - redis
        gnatsd  - gnatsd
        db      - postgres or none if working from sqlite
        rest    - the rest api
        web     - the web api

stop:
    Stops the connectordb instance running from the directory

`

const (
	//ConnectorDBConfigFileName is the file name to use for the configuration file in the database folder
	ConnectorDBConfigFileName = "cdb.ini"

	//DefaultFolderPermissions is The folder permissions to use when creating a database
	DefaultFolderPermissions = os.FileMode(0755)
)

var (
	createFlags = flag.NewFlagSet("create", flag.ExitOnError)

	// TODO change this in the future, but until we have the final thing ironed out, leave this as is for testing
	createUsernamePassword = createFlags.String("user", "admin:admin", "The user in username:password format")
	createEmail            = createFlags.String("email", "root@localhost", "The email address for the root user")
	createDbType 		   = createFlags.String("dbtype", "postgres", "The type of database to create.")

	startFlags = flag.NewFlagSet("start", flag.ExitOnError)

	// True if
	teardownDaemons = false
)

// The main entrypoint into connectordb
func main() {

	// we parse the flags here to make sure the usage will perform correctly
	flag.Parse()

	if len(flag.Args()) < 2 {
		fmt.Printf(ProgramUsage)
		flag.Usage()
		os.Exit(1)
	}

	// we use flag args just in case the user put flags before the params
	commandName := flag.Args()[0]
	processDirectory := flag.Args()[1]

	switch commandName {
	case "create":
		create(processDirectory)
	case "start":
		start(processDirectory)
	case "stop":
		stop(processDirectory)
	case "shell":
		fmt.Printf("Not yet implemented\n")
	default:
		fmt.Printf("Error: '%v' is not a valid command.\n", commandName)
		fmt.Printf(ProgramUsage)
		os.Exit(1)
	}
}


func waitForPortOpen(host string, port int) {

	hostPort := fmt.Sprintf("%s:%d", host, port)

	var err error

	log.Printf("Waiting for %v to open...\n", hostPort)


	_, err = net.Dial("tcp", hostPort)

	for err != nil {
		_, err = net.Dial("tcp", hostPort)
	}

	log.Printf("...%v is now open.\n", hostPort)

}

// Executes a command, redirecting the stdout and stderr to this program's output
func executeCommand(command string, args ...string) error {
	log.Printf(cmd2Str(command, args...))
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func cmd2Str(command string, args ...string) string {
	return fmt.Sprintf("%v %v", command, strings.Join(args, " "))
}

// Executes a command, redirecting the stdout and stderr to this program's output
func daemonizeCommand(logpidpath, command string, args ...string) error {
	// TODO setup infinite loop for restarting crashed processes

	log.Printf("Starting Daemon: %v\n", cmd2Str(command, args...))


	file, err := os.Create(logpidpath + ".log") // For read access.
	if err != nil {
		return err
	}

	go execCommandRedirect(file, file, logpidpath+".pid", command, args...)

	return nil
}

// Executes a command doing redirects as necessary
func execCommandRedirect(stdout, stderr *os.File, pidpath string, command string, args ...string) {
	for {
		log.Printf("Executing: %v\n", cmd2Str(command, args...))

		cmd := exec.Command(command, args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Start()

		pidfile, err := os.Create(pidpath) // For read access.
		if err == nil {
			if cmd.Process != nil {
				pidfile.WriteString(fmt.Sprintf("%v", cmd.Process.Pid))
			}
			pidfile.Close()
		} else {
			log.Printf("%v\n", err.Error())
		}
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

		cmdexit := make(chan bool)
		go func() {
			cmd.Wait()
		    cmdexit <- true
		}()

		teardownall := make(chan bool)
		go func() {
			for {
				time.Sleep(100 * time.Millisecond)
				if teardownDaemons {
					teardownall <- true
				}
			}
		}()

		select {
			case <-ch:
				log.Printf("Stopping process: %v\n", cmd2Str(command, args...))
				if cmd.Process != nil {
					cmd.Process.Signal(syscall.SIGTERM)
				}
				return
			case <- teardownall:
				log.Printf("Global teardown, stopping process: %v\n", cmd2Str(command, args...))
				if cmd.Process != nil {
					cmd.Process.Signal(syscall.SIGTERM)
				}
				return
			case <-cmdexit:
				log.Printf("ERROR: %v failed, restarting.\n", command)
		}
	}

	// TODO setup a global PID structure to track all pending and past operations
}

func createDeviceAndGetKey(udb *users.UserDatabase, user *users.User, devname string) (string, error) {
	err := udb.CreateDevice(devname, user.UserId)
	if err != nil {
		return "", err
	}

	dev, err := udb.ReadDeviceForUserByName(user.UserId, devname)
	if err != nil {
		return "", err
	}

	return dev.ApiKey, nil
}

// Handles an error by logging it fatally if it exists
func fatalHandleError(err error) {
	if err != nil {
		log.Println(err.Error())
		teardownAll()
		os.Exit(1)
	}
}

func appendLinesToFile(filepath string, lines ...string) error {
	appendchunk := strings.Join(lines, "\n")

	fd, err := os.OpenFile(filepath, os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return err
	}

	_, err = fd.WriteString(appendchunk)

	fd.Close()
	return err
}

func create(ProcessDir string) {
	createFlags.Parse(os.Args[2:])

	userPass := strings.Split(*createUsernamePassword, ":")
	if len(userPass) != 2 {
		log.Fatal("Username and password not given in format <username>:<password>\n")
		return
	}
	createUsername := userPass[0]
	createPassword := userPass[1]

	execFolder, _ := osext.ExecutableFolder()

	log.Printf("Initial Setup...\n")

	// Create the initial directory
	log.Printf("> Creating Directory\n")

	err := os.MkdirAll(ProcessDir, DefaultFolderPermissions)
	fatalHandleError(err)

	// Copy the config files over to the new folder
	log.Printf("> Copying Config\n")

	err = executeCommand("cp", filepath.Join(execFolder, "/config/gnatsd.conf"), filepath.Join(execFolder, "/config/redis.conf"), ProcessDir)
	fatalHandleError(err)


	databasePath := ""
	switch *createDbType {
	case "sqlite":
		databasePath = "connectordb.sqlite3"

		// because sqlite doesn't always like being started on a file that
		// doesn't exist
		executeCommand("touch", databasePath)

	default: // postgres or misconfigured

		// Init the postgres database
		log.Printf("Setting Up Postgres\n")

		postgresPath := filepath.Join(ProcessDir, "connectordb_psql")

		err = os.MkdirAll(postgresPath, DefaultFolderPermissions)
		fatalHandleError(err)

		err = executeCommand(dbutil.FindPostgresInit(), "-D", postgresPath)
		fatalHandleError(err)


		postgresPortString := fmt.Sprintf("%d", *config.PostgresPort)

		// Setup postgres.conf
		confPath := filepath.Join(postgresPath, "postgresql.conf")
		err = appendLinesToFile(confPath, "\n",
									"port = " + postgresPortString,
									"listen_addresses = 'localhost'",
									"unix_socket_directories = '/tmp'",
									"\n")

		fatalHandleError(err)

		startPostgres(ProcessDir)

		// Wait for the port to open
		time.Sleep(time.Second * 3)
		waitForPortOpen("localhost", *config.PostgresPort)

		// Create the initial database
		err = executeCommand(dbutil.FindPostgresPsql(), "-h", "localhost", "-p", postgresPortString, "-d", "postgres", "-c", "CREATE DATABASE connectordb;")
		fatalHandleError(err)

		databasePath = fmt.Sprintf("postgres://localhost:%d/connectordb?sslmode=disable", *config.PostgresPort)
	}

	// Setup the tables
	log.Println("Upgrading database")
	err = dbutil.UpgradeDatabase(databasePath, true)
	fatalHandleError(err)

	// Setup the admin user and two main devices
	log.Printf("Setting up the initial admin user.")
	db, driver, err := dbutil.OpenSqlDatabase(databasePath)
	fatalHandleError(err)


	var udb users.UserDatabase
	udb.InitUserDatabase(db, string(driver))

	// create the initial user
	err = udb.CreateUser(createUsername, *createEmail, createPassword)
	if err != nil {
		log.Fatal("user create failed:" + err.Error())
	}
	/*
		usr, err := udb.ReadUserByName(createUsername)
		if err != nil {
			log.Fatal("read user failed:" + err.Error())
		}
	*/

	log.Printf("Initializing local configuration.")

	// Setup config
	flag.Set("database.cxn_string", databasePath)
	flag.Set("user", "")

	// Dump config
	config := getConfig()


	cfgfile := filepath.Join(ProcessDir ,ConnectorDBConfigFileName)
	file, err := os.Create(cfgfile) // For read access.
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()
	file.WriteString(config)

	log.Println("Setup was successful, exiting.")

	teardownAll()
}

// stolen from iniflags
func quoteValue(v string) string {
	if !strings.ContainsAny(v, "\n#;") && strings.TrimSpace(v) == v {
		return v
	}
	v = strings.Replace(v, "\\", "\\\\", -1)
	v = strings.Replace(v, "\n", "\\n", -1)
	v = strings.Replace(v, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", v)
}

func escapeUsage(s string) string {
	return strings.Replace(s, "\n", "\n    # ", -1)
}

func getConfig() string {
	config := ""

	flag.VisitAll(func(f *flag.Flag) {
		if f.Name != "config" && f.Name != "dumpflags" {
			config += fmt.Sprintf("%s = %s  # %s\n", f.Name, quoteValue(f.Value.String()), escapeUsage(f.Usage))
		}
	})

	return config
}

func startPostgres(ProcessDir string) error {
	log.Println("Starting postgres")

	executablePath := dbutil.FindPostgres()
	dbPath := filepath.Join(ProcessDir, "connectordb_psql")
	port := fmt.Sprintf("%d", *config.PostgresPort)

	if executablePath == "" {
		log.Fatal("Could not find postgres path\n")
	}

	//return executeCommand(executablePath, "-p", "52592", "-D", dbPath)
	return daemonizeCommand("postgres", executablePath, "-p", port, "-D", dbPath)
}

func startGnatsd(ProcessDir string) error {
	log.Println("Starting gnatsd")

	execFolder, _ := osext.ExecutableFolder()
	binaryPath := filepath.Join(execFolder, "dep/gnatsd")
	configPath := filepath.Join(ProcessDir, "gnatsd.conf")

	return daemonizeCommand("gnatsd", binaryPath, "-c", configPath)
}

func startRedis(ProcessDir string) error {
	log.Println("Starting redis")

	configPath := filepath.Join(ProcessDir, "redis.conf")

	return daemonizeCommand("redis", "redis-server", configPath)
}

func start(ProcessDir string) {
	execFolder, _ := osext.ExecutableFolder()
	ProcessDir, _ = filepath.Abs(ProcessDir)

	cdbConfigPath := filepath.Join(ProcessDir, ConnectorDBConfigFileName)

	//Change to the directory
	err := os.Chdir(ProcessDir)
	if err != nil {
		log.Println("Could not change local directory: ", err)
	}

	fmt.Printf("Starting connectordb...\n\n")
	fmt.Printf("Exec Folder: %v\n", execFolder)
	fmt.Printf("DB Folder  : %v\n", ProcessDir)
	fmt.Printf("ini path   : %v\n\n\n", cdbConfigPath)

	// load configuration, first we start with the flags library so we can
	// specify the loading path...
	flag.Parse()
	flag.Set("config", cdbConfigPath) // the inipath for iniflags

	iniflags.Parse() // Now we setup the iniflags which handles the sighup stuff

	subSubcommand := ""
	if len(flag.Args()) >= 3 {
		subSubcommand = flag.Args()[2]
	}

	var redisNeeded, dbNeeded, webNeeded, restNeeded, gnatsdNeeded bool
	switch subSubcommand {
	case "servers":
		redisNeeded = true
		dbNeeded = true
		gnatsdNeeded = true

	case "redis":
		redisNeeded = true

	case "gnatsd":
		gnatsdNeeded = true

	case "db":
		dbNeeded = true

	case "rest":
		restNeeded = true

	case "web":
		restNeeded = true
		webNeeded = true

	default:
		redisNeeded = true
		dbNeeded = true
		webNeeded = true
		restNeeded = true
		gnatsdNeeded = true
	}

	// Now start all the services we need
	if gnatsdNeeded {
		startGnatsd(ProcessDir)
	}

	if redisNeeded {
		startRedis(ProcessDir)
	}

	if dbNeeded && !dbutil.UriIsSqlite(*config.DatabaseConnection) {
		startPostgres(ProcessDir)
	}

	if webNeeded || restNeeded {
		time.Sleep(time.Second * 3)

		// start api + webservice
		var err error
		//db, err := streamdb.Open(*config.DatabaseConnection, *config.RedisConnection, *config.MessageConnection)
		//NOT dealing with this shit right now - null pointer here in config.
		db, err := streamdb.Open("postgres://localhost:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")

		if err != nil {
			panic(err.Error())
		}

		defer db.Close()
		log.Println("Running REST server")
		r := rest.Router(db, nil)
		http.Handle("/", r)

		serveraddr := fmt.Sprintf(":%d", *config.WebPort)
		err = http.ListenAndServe(serveraddr, nil)
		log.Fatal(err)
	}
	teardownAll()
}

func stop(ProcessDir string) {
	fmt.Printf("Stopping connectordb...\n")
	//Change to the directory
	err := os.Chdir(ProcessDir)
	if err != nil {
		log.Println("Could not change local directory: ", err)
	}

	postgrespid := filepath.Join("connectordb_psql", "postmaster.pid")
	exec.Command("bash", "-c", fmt.Sprintf("kill `head -n 1 %v`", postgrespid)).Run()

	pids, err := filepath.Glob("*.pid")
	for _, pid := range pids {
		log.Println("Kill:", pid)
		err = exec.Command("bash", "-c", fmt.Sprintf("kill `head -n 1 %v`", pid)).Run()
		if err != nil {
			log.Println(err)
		}
	}
}

func teardownAll(){
	log.Printf("Sending system teardown request; trying to kill all threads and processes.\n")
	teardownDaemons = true
	time.Sleep(3 * time.Second)
}
