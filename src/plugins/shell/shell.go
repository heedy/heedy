package shell

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"streamdb"
	"plugins"
)

const (
	Reset = "\x1b[0m"
	Bold = "\x1b[1m"
	Black = "\x1b[30m"
 	Red = "\x1b[31m"
	Green = "\x1b[32m"
	Yellow = "\x1b[33m"
	Blue  = "\x1b[34m"
	Magenta  = "\x1b[35m"
	Cyan  = "\x1b[36m"
	White = "\x1b[37m"

	Password = "\x1b[30;40m" // black on black

	cdbshell = `
   ___                      _           ___  ___   ___ _        _ _
  / __|___ _ _  _ _  ___ __| |_ ___ _ _|   \| _ ) / __| |_  ___| | |
 | (__/ _ \ ' \| ' \/ -_) _|  _/ _ \ '_| |) | _ \ \__ \ ' \/ -_) | |
  \___\___/_||_|_||_\___\__|\__\___/_| |___/|___/ |___/_||_\___|_|_|
`
)

func init() {
	// do some sweet plugin registration!
	plugins.Register("shell", usage, startShellExec)
}

func startShellExec(sdb *streamdb.Database, args []string) error {
	StartShell(sdb)
	return nil
}

func usage() {
	fmt.Println(`shell: runs an interactive shell for connectordb

    Currently only basic utilities are supported, but more will come soon.
    This is the command you want to use to add/modify/delete users, view the
    health of your system and/or do administrative tasks.

    In the future it will be possible to script the shell to make administration
    easier.
`)
}



func StartShell(sdb *streamdb.Database) {
	s := CreateShell(sdb)
	s.Cls()
	s.Motd()
	s.Repl()
}

// The shell we're operating under
type Shell struct {
	VersionString string
	CopyrightString string
	running bool
	commands []ShellCommand
	host string
	reader *bufio.Reader
	sdb *streamdb.Database
	operator streamdb.Operator
	operatorName string // can be changed when we do a su
}

func (s *Shell) Repl() {
	for s.running {
		fmt.Printf(s.GetPrompt())
		text := s.ReadLine()
		s.RunCommand(text)

	}
}

func (s *Shell) RunCommand(cmdstring string) {
	cmdstring = strings.TrimSpace(cmdstring)
	command := strings.Split(cmdstring, " ")
	if len( command ) == 0 {
		return
	}

	for _, cmd := range(s.commands) {
		if cmd.Name() == command[0] {
			cmd.Execute(s, command)
			return
		}
	}

	fmt.Printf("Command '%v' not found, use 'help' to list available commands\n", cmdstring)
}

func CreateShell(sdb *streamdb.Database) *Shell {
	var s Shell
	s.VersionString = "ConnectorDB Shell v 1.0"
	s.CopyrightString = "Copyright Joseph Lewis & Daniel Kumor 2015"
	s.running = true
	s.commands = []ShellCommand{
		Help{},
		Exit{},
		Clear{},
		GrantAdmin{},
		RevokeAdmin{},
		AddUser{},
		ListUsers{},
		Cat{},
		Su{},
		ListDevices{}}
	s.host, _ = os.Hostname()
	s.reader = bufio.NewReader(os.Stdin)
	s.sdb = sdb
	s.operator = sdb.GetAdminOperator()
	s.operatorName = "ConnectorDB"
	return &s
}

func (s *Shell) GetPrompt() string {
	return Bold + Magenta + s.operatorName + White + "@" + Blue + s.host + White + ":" + Cyan + "~" + White + "> " + Reset
}

// Prints a seperator
func (s *Shell) Seperator() {

	for i := 0; i < 80; i++{
		fmt.Printf("-")
	}

	fmt.Printf("\n")
}

// Clears the screen (on VT100 terminals)
func (s *Shell) Cls() {
	fmt.Printf("\033[H\033[2J\n")
}

// Prints the message of the day
// In the future, we'll use this like UNIX does as a general alert system
func (s *Shell) Motd() {
	fmt.Printf(Blue + cdbshell + Reset)
	fmt.Println()
	fmt.Printf("%v\n", s.VersionString)
	fmt.Printf("%v\n\n", s.CopyrightString)
}

// Reads a line of input from the shell
func (s *Shell) ReadLine() string {
	str, _ := s.reader.ReadString('\n')
	return strings.TrimSpace(str)
}

// Prints an error if it exists. Returns true if printed, false if not
func (s *Shell) PrintError(err error) bool {
	if err != nil {
		fmt.Printf(Red + "Error: %v\n" + Reset, err.Error())
	}

	return err != nil
}

// The ShellCommand is an internal command within our internal shell.
type ShellCommand interface {
		// Returns the help string associated with this command.
        Help() string

		// Returns the help for a specific command
		Usage() string

		// Execute the command with the given arguments
		Execute(shell *Shell, args []string)

		// Returns the name of this shell command, should be all lower case
		Name() string
}


// The help command
type Help struct {
}

func (h Help) Help() string {
	return "Prints this dialog"
}

func (h Help) Usage() string {
	return `Displays help information about the built in commands.

	Usage: help [commandname]

	The optional command name will show more detailed information about a given
	command.
`
}

func (h Help) Execute(shell *Shell, args []string) {
	if len(args) == 2 {
		for _, cmd := range(shell.commands) {
			if cmd.Name() == args[1] {
				fmt.Println(Bold)
				fmt.Printf("%s Help\n" + Reset, args[1])
				fmt.Println("")
				fmt.Printf(cmd.Usage())
				return
			}
		}
		fmt.Printf(Red + "%s not found, listing known commands:\n" + Reset, args[1])
	}

	fmt.Println(Bold)
	fmt.Printf("ConnectorDB Shell Help\n" + Reset)
	fmt.Println("")

	for _, cmd := range(shell.commands) {
		fmt.Printf("%v\t- %v\n", cmd.Name(), cmd.Help())
	}
	fmt.Println("")
	fmt.Println("Use 'help [commandname]' to show help for a specific command.")
}

func (h Help) Name() string {
	return "help"
}


// The Exit command
type Exit struct {
}

func (h Exit) Help() string {
	return "Quits the interactive shell"
}

func (h Exit) Usage() string {
	return h.Help()
}

func (h Exit) Execute(shell *Shell, args []string) {
	fmt.Printf("exit\n")
	shell.running = false
}

func (h Exit) Name() string {
	return "exit"
}

// The clear command
type Clear struct {
}

func (h Clear) Help() string {
	return "Clears the screen"
}

func (h Clear) Usage() string {
	return ""
}

func (h Clear) Execute(shell *Shell, args []string) {
	fmt.Println(Reset) // clear the shell's color problems if any
	shell.Cls()
}

func (h Clear) Name() string {
	return "clear"
}

// The clear command
type AddUser struct {
}

func (h AddUser) Help() string {
	return "Creates a new user"
}

func (h AddUser) Usage() string {
	return ""
}

func (h AddUser) Execute(shell *Shell, args []string) {
	// TODO grant admin
	fmt.Print("Enter the name for the new user: ")
	name := shell.ReadLine()

	fmt.Print("Enter the email for the new user: ")
	email := shell.ReadLine()


	// Do the password check
	passdiff := true
	pass1 := ""
	for passdiff {
		fmt.Println("Enter password for new user:")
		fmt.Print(Password)
		pass1 = shell.ReadLine()
		fmt.Println(Reset + "Re-enter password:" + Black)
		pass2 := shell.ReadLine()
		fmt.Print(Reset)

		if pass1 == pass2 {
			passdiff = false
		} else {
			fmt.Println("Passwords did not match, type 'yes' to try again")
			decision := shell.ReadLine()
			if decision != "yes" {
				return
			}
		}
	}

	fmt.Printf("Creating User %v at %v\n", name, email)

	err := shell.operator.CreateUser(name, email, pass1)
	if err != nil {
		fmt.Printf(Red + "Error: %v\n" + Reset, err.Error())
	}
}

func (h AddUser) Name() string {
	return "adduser"
}
