package shell

import (
	"bufio"
	"config"
	"connectordb"
	"connectordb/operator"
	"connectordb/users"
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

var cfg = config.NewConfiguration()

//Sets the configuration
func SetConfiguration(c *config.Configuration) {
	cfg = c
}

func startShellExec(sdb *connectordb.Database, args []string) error {
	if len(args) == 0 {
		StartShell(sdb)
	}

	s := CreateShell(sdb)
	s.execCommand(args...)
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

func StartShell(sdb *connectordb.Database) {
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
	host            string
	reader          *bufio.Reader
	sdb             *connectordb.Database
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

	s.execCommand(command...)
}

func (s *Shell) execCommand(command ...string) uint8 {
	if len(command) == 0 {
		return 1
	}

	for _, cmd := range allCommands {
		if cmd.name == command[0] {

			return cmd.main(s, command)
		}
	}

	fmt.Printf("Command '%v' not found, use 'help' to list available commands\n", command[0])
	return 1
}

func CreateShell(sdb *connectordb.Database) *Shell {
	var s Shell
	s.VersionString = "ConnectorDB Shell v" + connectordb.Version
	s.CopyrightString = "Copyright Joseph Lewis & Daniel Kumor 2015"
	s.running = true
	s.host, _ = os.Hostname()
	s.reader = bufio.NewReader(os.Stdin)
	s.sdb = sdb
	s.operator = operator.NewOperator(sdb)
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

// Prints a question then returns ther user's answer
func (s *Shell) ReadAnswer(question string) string {
	fmt.Print(question)
	return s.ReadLine()
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
		s.PrintErrorText("Error: %v", err.Error())
	}

	return err != nil
}

// Reads the user, device and stream at a path
func (s *Shell) ReadPath(path string) (usr *users.User, dev *users.Device, stream *users.Stream) {
	usr, _ = s.operator.ReadUser(path)
	dev, _ = s.operator.ReadDevice(path)
	stream, _ = s.operator.ReadStream(path)

	return usr, dev, stream
}

// prepends the current working place to the give path
func (s *Shell) ResolvePath(path string) string {
	return s.pwd + path
}

// Prints something in title text
func (s *Shell) PrintTitle(title string) {
	fmt.Printf(Bold + title + Reset + "\n\n")
}

// Prints something in title text
func (s *Shell) PrintErrorText(format string, args ...interface{}) {
	fmt.Printf(Red+format+Reset, args...)
	fmt.Println("")
}
