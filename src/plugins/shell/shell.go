package shell

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"streamdb"
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
   ___                      _           ___  ___   ___ _        _ _   _   __
  / __|___ _ _  _ _  ___ __| |_ ___ _ _|   \| _ ) / __| |_  ___| | | / | /  \
 | (__/ _ \ ' \| ' \/ -_) _|  _/ _ \ '_| |) | _ \ \__ \ ' \/ -_) | | | || () |
  \___\___/_||_|_||_\___\__|\__\___/_| |___/|___/ |___/_||_\___|_|_| |_(_)__/
`
)



func StartShell(sdb *streamdb.Database) {
	s := CreateShell()
	s.Cls()
	s.Motd()

	for s.running {
		fmt.Printf(s.GetPrompt())
		text := s.ReadLine()
		s.RunCommand(text)

	}
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
	s.commands = []ShellCommand{Help{}, Exit{}, Clear{}, GrantAdmin{}, AddUser{}}
	s.host, _ = os.Hostname()
	s.reader = bufio.NewReader(os.Stdin)
	s.sdb = sdb
	return &s
}

func (s *Shell) GetPrompt() string {
	return Bold + Magenta + "ConnectorDB" + White + "@" + Blue + s.host + White + ":" + Cyan + "~" + White + "> " + Reset
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

// The ShellCommand is an internal command within our internal shell.
type ShellCommand interface {
		// Returns the help string associated with this command.
        Help() string

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

func (h Help) Execute(shell *Shell, args []string) {
	fmt.Println(Bold)
	fmt.Printf("ConnectorDB Shell Help\n" + Reset)
	fmt.Println("")

	for _, cmd := range(shell.commands) {
		fmt.Printf("%v\t- %v\n", cmd.Name(), cmd.Help())
	}
	fmt.Println("")
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

func (h Clear) Execute(shell *Shell, args []string) {
	shell.Cls()
}

func (h Clear) Name() string {
	return "clear"
}



// The clear command
type GrantAdmin struct {
}

func (h GrantAdmin) Help() string {
	return "Grants admin to a user: 'grantadmin username'"
}

func (h GrantAdmin) Execute(shell *Shell, args []string) {
	// TODO grant admin
	output := ""
	if len(args) < 2 {
		output = Red + "Must supply a name" + Reset
	} else {
		output = Green + "Granted admin to: " + args[1] + Reset
	}
	fmt.Println(output)
}

func (h GrantAdmin) Name() string {
	return "mkadmin"
}



// The clear command
type AddUser struct {
}

func (h AddUser) Help() string {
	return "Creates a new user"
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

	err := sdb.CreateUser(name, email, pass1)
	if err != nil {
		fmt.Printf(Red + "Error: %v\n" + Reset, err.Error())
	}
}

func (h AddUser) Name() string {
	return "adduser"
}
