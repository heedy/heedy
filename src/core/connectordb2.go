package main

import (
    "flag"
    "fmt"
    "os"
    "os/exec"
    "streamdb/dbutil"
    "log"
    "net"
    "time"
    "strings"
    "streamdb/users"
    "github.com/vharitonsky/iniflags"
    "streamdb/config"
    "streamdb"
	"net/http"
	//"github.com/gorilla/mux"
    "plugins/rest"
    "path/filepath"
    "github.com/kardianos/osext"
)

const PROG_USAGE = `

Usage: connectordb (create|start|stop|shell) directory

directory - the directory holding the configuration and the data for connectordb

create:
    Sets up a new connectordb instance at the given directory.

start:
    Starts connectordb and needed processes running from the given connectordb
    instance directory.

stop:
    Stops the connectordb instance running from the directory

`

const DEFAULT_FOLDER_PERMISSIONS = os.FileMode(0755)

const (
    CONNECTORDB_CONFIG_FILE_NAME = "cdb.ini"
)

var (
    create_flags = flag.NewFlagSet("create", flag.ContinueOnError)
    create_user  = flag.String("username", "admin", "The admin user name")
    create_pass  = flag.String("password", "admin", "The admin default password")
    create_email = flag.String("email", "root@localhost", "The admin email address")
)

// The main entrypoint into connectordb
func main() {
	flag.Parse()

    if len( flag.Args() ) < 2 {
        fmt.Printf(PROG_USAGE)
        flag.Usage()
        os.Exit(1)
    }

    command_name := flag.Args()[0]
    process_directory := flag.Args()[1]

    switch command_name {
        case "create":
            create(process_directory)
        case "start":
            start(process_directory)
        case "stop":
            stop(process_directory)
        case "shell":
            fmt.Printf("Not yet implemented\n")
        default:
            fmt.Printf("Error: '%v' is not a valid command.\n", command_name)
            fmt.Printf(PROG_USAGE)
            os.Exit(1)
    }
}

func waitForPortOpen(host_port string) {
    var err error

    _, err = net.Dial("tcp", host_port)

    for err != nil {
        _, err = net.Dial("tcp", host_port)
    }
}


// Executes a command, redirecting the stdout and stderr to this program's output
func executeCommand(command string, args ...string) error {

    return execCommandRedirect(os.Stdout, os.Stderr, command, args...)
}



// Executes a command, redirecting the stdout and stderr to this program's output
func daemonizeCommand(logpath, command string, args ...string) error {

    file, err := os.Create(logpath) // For read access.
    if err != nil {
    	return err
    }

    go execCommandRedirect(file, file, command, args...)

    return nil
}

// Executes a command doing redirects as necessary
func execCommandRedirect(stdout, stderr *os.File, command string, args ...string) error {
	cmd := exec.Command(command, args...)

    cmd.Stdout = stdout
    cmd.Stderr = stderr

    return cmd.Run()
}


func create_dev_and_get_key(udb *users.UserDatabase, user *users.User, devname string) (string, error) {
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


func create(ProcessDir string) {
    create_flags.Parse(os.Args)

    exec_folder, _ := osext.ExecutableFolder()

    log.Printf("Initial Setup...\n")

    // Create the initial directory
    log.Printf("> Creating Directory\n")


    if err := os.MkdirAll(ProcessDir, DEFAULT_FOLDER_PERMISSIONS); err != nil {
        log.Fatal(err.Error())
    }

    // Copy the config files over to the new folder
    log.Printf("> Copying Config\n")

    err := executeCommand("cp", exec_folder + "/config/gnatsd.conf", exec_folder + "/config/redis.conf", ProcessDir)

    if err != nil {
        log.Fatal(err.Error())
    }

    database_path := ""
    dbtype := "sqlite"
    switch dbtype {
        case "sqlite":
            database_path = ProcessDir + "/connectordb.sqlite3"
            executeCommand("touch", database_path)

        default: // postgres or misconfigured

            // Init the postgres database
            fmt.Printf("> Setting Up Postgres\n")

            database_setup_dir := ProcessDir + "/connectordb_psql"

            err = executeCommand("bash", exec_folder + "/config/runpostgres", "setup", database_setup_dir)

            if err != nil {
                log.Fatal(err.Error())
            }

            err = daemonizeCommand("bash", exec_folder + "/config/runpostgres", "run", database_setup_dir)
            if err != nil {
                log.Fatal(err.Error())
            }

            database_path = "postgres://localhost:52592/connectordb?sslmode=disable"
            log.Printf("Waiting for port to open.")
            waitForPortOpen("localhost:52592")
            time.Sleep(time.Second * 5)
    }


    // Setup the tables
    err = dbutil.UpgradeDatabase(database_path, true)

    if err != nil {
        log.Fatal(err.Error())
    }

    // Setup the admin user and two main devices
    log.Printf("Setting up the initial admin user.")
    db, driver, err := dbutil.OpenSqlDatabase(database_path)

    if err != nil {
        log.Fatal(err.Error())
    }

    var udb users.UserDatabase
    udb.InitUserDatabase(db, string(driver))

    // create the initial user
    err = udb.CreateUser(*create_user, *create_email, *create_pass)
    if err != nil {
        log.Fatal(err.Error())
    }

    usr, err := udb.ReadUserByName(*create_user)
    if err != nil {
        log.Fatal(err.Error())
    }


    restkey, err := create_dev_and_get_key(&udb ,usr, "REST Api")
    if err != nil {
        log.Fatal(err.Error())
    }

    webkey, err := create_dev_and_get_key(&udb ,usr, "Website")
    if err != nil {
        log.Fatal(err.Error())
    }



    // Setup config
    flag.Set("database.cxn_string", database_path)
    flag.Set("web.api.key", restkey)
    flag.Set("web.http.key", webkey)
    flag.Set("username", "")
    flag.Set("password", "")


    // Dump config

    config := getConfig()

    file, err := os.Create(ProcessDir + "/" + CONNECTORDB_CONFIG_FILE_NAME) // For read access.
    if err != nil {
    	log.Fatal(err.Error())
    }

    defer file.Close()
    file.WriteString(config)

    log.Println("Finished all setup, exiting.")
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

func start_gnatsd(ProcessDir string) {
    exec_folder, _ := osext.ExecutableFolder()

    log.Println("Starting gnatsd")
    daemonizeCommand(ProcessDir + "/gnatsd.log", exec_folder + "/dep/gnatsd", "-c", ProcessDir + "/gnatsd.conf")
}

func start_redis(ProcessDir string) {
    log.Println("Starting redis")
    daemonizeCommand(ProcessDir + "/redis.log", "redis-server", ProcessDir + "/redis.conf")
}

func start(ProcessDir string) {
    fmt.Printf("Starting connectordb from '%v'\n", ProcessDir)

    // load configuration, first we start with the flags library so we can
    // specify the loading path...

    flag.Parse()
    cdb_config_path := ProcessDir + "/" + CONNECTORDB_CONFIG_FILE_NAME
    cdb_config_path, _ = filepath.Abs(cdb_config_path)

    flag.Set("config", cdb_config_path) // the inipath for iniflags

    iniflags.Parse() // Now we setup the iniflags which handles the sighup stuff


    // Start other services
    start_gnatsd(ProcessDir)
    start_redis(ProcessDir)

    time.Sleep(time.Second * 3)

    // start api + webservice
	var err error
	db, err := streamdb.Open(*config.DatabaseConnection, *config.RedisConnection, *config.MessageConnection)

	if err != nil {
		log.Println("Cannot open StreamDB")
		panic(err.Error())
	}

	defer db.Close()
    r := rest.Router(db, nil)
	http.Handle("/", r)

    serveraddr := fmt.Sprintf(":%d", *config.WebPort)
    err = http.ListenAndServe(serveraddr, nil)
	log.Fatal(err)
}

func stop(ProcessDir string) {
    fmt.Printf("Stopping connectordb...\n")
}
