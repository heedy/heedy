package shell

import "fmt"

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
	fmt.Println(Reset)
	shell.Cls()
}

func (h Clear) Name() string {
	return "clear"
}
