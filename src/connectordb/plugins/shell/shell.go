package shell

import (
	"bufio"
	"connectordb/plugins"
	"connectordb/streamdb"
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"
	"fmt"
	"os"
	"strings"
)

const (
	Reset   = "\x1b[0m"
	Bold    = "\x1b[1m"
	Black   = "\x1b[30m"
	Red     = "\x1b[31m"
	Green   = "\x1b[32m"
	Yellow  = "\x1b[33m"
	Blue    = "\x1b[34m"
	Magenta = "\x1b[35m"
	Cyan    = "\x1b[36m"
	White   = "\x1b[37m"

	Password      = "\033[30;40m" // black on black
	ClearLastLine = "\033[1A\033[2K\033[1A"

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
	VersionString   string
	CopyrightString string
	running         bool
	commands        []ShellCommand
	host            string
	reader          *bufio.Reader
	sdb             *streamdb.Database
	operator        operator.Operator
	operatorName    string // can be changed when we do a su
	pwd             string // the present working directory of path commands
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
	if len(command) == 0 {
		return
	}

	for _, cmd := range s.commands {
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
		ListDevices{},
		Passwd{},
		Rm{},
		Ls{}}
	s.host, _ = os.Hostname()
	s.reader = bufio.NewReader(os.Stdin)
	s.sdb = sdb
	s.operator = sdb.Operator
	s.operatorName = "ConnectorDB"
	s.pwd = ""
	return &s
}

func (s *Shell) GetPrompt() string {
	return Bold + Magenta + s.operatorName + White + "@" + Blue + s.host + White + ":" + Cyan + "~" + White + "> " + Reset
}

// Prints a seperator
func (s *Shell) Seperator() {

	for i := 0; i < 80; i++ {
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

// Reads a password from the command line
func (s *Shell) ReadPassword() string {
	fmt.Printf("Password: " + Password)
	passwd := s.ReadLine()

	fmt.Println(Reset + ClearLastLine)
	return passwd
}

// Reads a password from the command line, return will be blank on failure
func (s *Shell) ReadRepeatPassword() string {
	fmt.Printf("Password: " + Password)
	passwd := s.ReadLine()
	fmt.Println(Reset + ClearLastLine)

	fmt.Printf("Repeat Password: " + Password)
	passwd2 := s.ReadLine()
	fmt.Println(Reset + ClearLastLine)

	if passwd != passwd2 {
		fmt.Println(Yellow + "Passwords did not match" + Reset)
		return ""
	}

	return passwd
}

// Prints an error if it exists. Returns true if printed, false if not
func (s *Shell) PrintError(err error) bool {
	if err != nil {
		fmt.Printf(Red+"Error: %v\n"+Reset, err.Error())
	}

	return err != nil
}

// Reads the user, device and stream at a path
func (s *Shell) ReadPath(path string) (usr *users.User, dev *users.Device, stream *operator.Stream) {
	usr, _ = s.operator.ReadUser(path)
	dev, _ = s.operator.ReadDevice(path)
	stream, _ = s.operator.ReadStream(path)

	return usr, dev, stream
}

// prepends the current working place to the give path
func (s *Shell) ResolvePath(path string) string {
	return s.pwd + path
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
		for _, cmd := range shell.commands {
			if cmd.Name() == args[1] {
				fmt.Println(Bold)
				fmt.Printf("%s Help\n"+Reset, args[1])
				fmt.Println("")
				fmt.Printf(cmd.Usage())
				return
			}
		}
		fmt.Printf(Red+"%s not found, listing known commands:\n"+Reset, args[1])
	}

	fmt.Println(Bold)
	fmt.Printf("ConnectorDB Shell Help\n" + Reset)
	fmt.Println("")

	for _, cmd := range shell.commands {
		fmt.Printf("%v\t- %v\n", cmd.Name(), cmd.Help())
	}
	fmt.Println("")
	fmt.Println("Use 'help [commandname]' to show help for a specific command.")
}

func (h Help) Name() string {
	return "help"
}
