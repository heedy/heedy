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
    "plugins/rest"
    "path/filepath"
    "github.com/kardianos/osext"
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

stop:
    Stops the connectordb instance running from the directory

`

//DefaultFolderPermissions is The folder permissions to use when creating a database
const DefaultFolderPermissions = os.FileMode(0755)

const (
    //ConnectorDBConfigFileName is the file name to use for the configuration file in the database folder
    ConnectorDBConfigFileName = "cdb.ini"
)

var (
    createFlags = flag.NewFlagSet("create", flag.ContinueOnError)
    createUsernamePassword  = createFlags.String("user", "", "The user in username:password format")
    createEmail = createFlags.String("email", "root@localhost", "The email address for the root user")
    createDbType = createFlags.String("dbtype","postgres","The type of database to create.")

    startFlags = flag.NewFlagSet("start",flag.ContinueOnError)
    startNoComponents = startFlags.Bool("databaseonly",false,"Only start the background servers, and not the interfaces")

)

// The main entrypoint into connectordb
func main() {

    if len(os.Args) < 3 {
        fmt.Printf(ProgramUsage)
        flag.Usage()
        os.Exit(1)
    }

    commandName := os.Args[1]
    processDirectory := os.Args[2]

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

func waitForPortOpen(hostPort string) {
    var err error

    _, err = net.Dial("tcp", hostPort)

    for err != nil {
        _, err = net.Dial("tcp", hostPort)
    }
}


// Executes a command, redirecting the stdout and stderr to this program's output
func executeCommand(command string, args ...string) error {

    cmd,_ :=  execCommandRedirect(os.Stdout, os.Stderr, command, args...)
    return cmd.Wait()
}



// Executes a command, redirecting the stdout and stderr to this program's output
func daemonizeCommand(logpath, command string, args ...string) (*exec.Cmd,error) {

    file, err := os.Create(logpath) // For read access.
    if err != nil {
    	return nil,err
    }

    return execCommandRedirect(file, file, command, args...)
}

// Executes a command doing redirects as necessary
func execCommandRedirect(stdout, stderr *os.File, command string, args ...string) (*exec.Cmd,error) {
	cmd := exec.Command(command, args...)

    cmd.Stdout = stdout
    cmd.Stderr = stderr

    return cmd,cmd.Start()
}

//TODO: Shouldn't this be part of streamdb? We should not need to use timebatchdb or userdb at all.
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


func create(ProcessDir string) {
    createFlags.Parse(os.Args[3:])

    userPass := strings.Split(*createUsernamePassword,":")
    if (len(userPass)!=2) {
        log.Fatal("Username and password not given in format <username>:<password>\n")
        return
    }
    createUsername := userPass[0]
    createPassword := userPass[1]

    execFolder, _ := osext.ExecutableFolder()

    log.Printf("Initial Setup...\n")

    // Create the initial directory
    log.Printf("> Creating Directory\n")


    if err := os.MkdirAll(ProcessDir, DefaultFolderPermissions); err != nil {
        log.Fatal(err.Error())
    }

    // Copy the config files over to the new folder
    log.Printf("> Copying Config\n")

    err := executeCommand("cp", execFolder + "/config/gnatsd.conf", execFolder + "/config/redis.conf", ProcessDir)

    if err != nil {
        log.Fatal(err.Error())
    }

    databasePath := ""
    dbtype := *createDbType
    var dbcmd *exec.Cmd
    switch dbtype {
        case "sqlite":
            databasePath = filepath.Join(ProcessDir,"connectordb.sqlite3")
            executeCommand("touch", databasePath)

        default: // postgres or misconfigured

            // Init the postgres database
            log.Printf("Setting Up Postgres\n")

            postgresPath := filepath.Join(ProcessDir,"connectordb_psql")

            err = executeCommand("bash", filepath.Join(execFolder,"config/runpostgres"), "setup", postgresPath)

            if err != nil {
                log.Fatal(err.Error())
            }

            databasePath = "postgres://localhost:52592/connectordb?sslmode=disable"

            time.Sleep(time.Second * 3)
            dbcmd,_ = startPostgres(ProcessDir)
            log.Printf("Waiting for port to open.")

            waitForPortOpen("localhost:52592")
            time.Sleep(time.Second * 2)
    }


    // Setup the tables
    err = dbutil.UpgradeDatabase(databasePath, true)

    if err != nil {
        if dbcmd!=nil {
            dbcmd.Process.Kill()
        }
        log.Fatal("Upgrade failed:"+err.Error())
    }

    // Setup the admin user and two main devices
    log.Printf("Setting up the initial admin user.")
    db, driver, err := dbutil.OpenSqlDatabase(databasePath)

    if err != nil {
        if dbcmd!=nil {
            dbcmd.Process.Kill()
        }
        log.Fatal("setup failed:"+err.Error())
    }

    var udb users.UserDatabase
    udb.InitUserDatabase(db, string(driver))

    // create the initial user
    err = udb.CreateUser(createUsername, *createEmail, createPassword)
    if err != nil {
        if dbcmd!=nil {
            dbcmd.Process.Kill()
        }
        log.Fatal("create failed:"+err.Error())
    }

    usr, err := udb.ReadUserByName(createUsername)
    if err != nil {
        if dbcmd!=nil {
            dbcmd.Process.Kill()
        }
        log.Fatal("read failed:"+err.Error())
    }

    //dkumor: Since the rest api and website will have streamdb compiled in, they should not need their own devices...
    //what do you think, @josephlewis42?
    restkey, err := createDeviceAndGetKey(&udb ,usr, "rest")
    if err != nil {
        if dbcmd!=nil {
            dbcmd.Process.Kill()
        }
        log.Fatal("create failed:"+err.Error())
    }

    webkey, err := createDeviceAndGetKey(&udb ,usr, "website")
    if err != nil {
        if dbcmd!=nil {
            dbcmd.Process.Kill()
        }
        log.Fatal("create failed:"+err.Error())
    }



    // Setup config
    flag.Set("database.cxn_string", databasePath)
    flag.Set("web.api.key", restkey)
    flag.Set("web.http.key", webkey)
    flag.Set("username", "")
    flag.Set("password", "")


    // Dump config

    config := getConfig()

    file, err := os.Create(ProcessDir + "/" + ConnectorDBConfigFileName) // For read access.
    if err != nil {
    	log.Fatal(err.Error())
    }

    defer file.Close()
    file.WriteString(config)

    log.Println("Finished all setup, exiting.")
    if dbcmd!=nil {
        dbcmd.Process.Kill()
    }
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

func startPostgres(ProcessDir string) (*exec.Cmd,error) {
    execFolder, _ := osext.ExecutableFolder()
    log.Println("Starting postgres")

    return daemonizeCommand(filepath.Join(ProcessDir,"postgres.log"),"bash", filepath.Join(execFolder,"config/runpostgres"), "run", filepath.Join(ProcessDir,"connectordb_psql"))
}

func startGnatsd(ProcessDir string) (*exec.Cmd,error) {
    execFolder, _ := osext.ExecutableFolder()

    log.Println("Starting gnatsd")
    return daemonizeCommand(filepath.Join(ProcessDir,"gnatsd.log"), filepath.Join(execFolder,"dep/gnatsd"), "-c", filepath.Join(ProcessDir,"gnatsd.conf"))
}

func startRedis(ProcessDir string) (*exec.Cmd,error) {
    log.Println("Starting redis")
    return daemonizeCommand(filepath.Join(ProcessDir,"redis.log"), "redis-server", filepath.Join(ProcessDir,"redis.conf"))
}

func start(ProcessDir string) {
    execFolder, _ := osext.ExecutableFolder()
    ProcessDir, _ = filepath.Abs(ProcessDir)

    cdbConfigPath := filepath.Join(ProcessDir,ConnectorDBConfigFileName)

    fmt.Printf("Starting connectordb...\n\n")
    fmt.Printf("Exec Folder: %v\n", execFolder)
    fmt.Printf("DB Folder  : %v\n", ProcessDir)
    fmt.Printf("ini path   : %v\n\n\n", cdbConfigPath)

    // load configuration, first we start with the flags library so we can
    // specify the loading path...
    flag.Parse()
    flag.Set("config", cdbConfigPath) // the inipath for iniflags

    iniflags.Parse() // Now we setup the iniflags which handles the sighup stuff


    // Start other services
    startPostgres(ProcessDir)
    startGnatsd(ProcessDir)
    startRedis(ProcessDir)

    time.Sleep(time.Second * 3)

    // start api + webservice
	var err error
	db, err := streamdb.Open(*config.DatabaseConnection, *config.RedisConnection, *config.MessageConnection)

	if err != nil {
		log.Println("Cannot open StreamDB")
		panic(err.Error())
	}

	defer db.Close()
    if (!*startNoComponents) {
        log.Println("Running REST server")
        r := rest.Router(db, nil)
    	http.Handle("/", r)

        serveraddr := fmt.Sprintf(":%d", *config.WebPort)
        err = http.ListenAndServe(serveraddr, nil)
    	log.Fatal(err)
    }
}

func stop(ProcessDir string) {
    fmt.Printf("Stopping connectordb...\n")
}
