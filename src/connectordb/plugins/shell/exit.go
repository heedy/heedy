package shell

import "fmt"

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
